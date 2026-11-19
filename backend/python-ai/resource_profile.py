"""
AuditSphere — Adaptive AI Training Resource Profile & Tier Selector.

Automatically detects host / container CPU & RAM capacity (including Docker cgroup v1/v2 limits)
and allows runtime override via environment variables on every job evaluation without requiring
code changes or server restarts.

Default profile for 2 vCPU / 8 GB RAM:
- AI_TRAIN_MAX_CONCURRENT = 1 (leaves 1 vCPU for API inference)
- AI_TRAIN_THREADS = 1
- AI_TRAIN_MAX_RAM_MB = 3000 (40% of 8 GB)
- AI_TRAIN_MIN_FREE_RAM_MB = 1500
- BATCH_ROWS = 5000
- RETRAIN_AUTO_INTERVAL = 60s
- RETRAIN_DEBOUNCE_SECONDS = 60s

Automatically scales up when deployed on larger servers (e.g. 8 vCPU, 16 vCPU, 64 GB RAM)
and selects training Tier (T1 In-Memory / T2 Out-of-Core Batch / T3 Stratified Reservoir Sampling)
based on dataset row count and RAM budget.
"""
import os
import math
from typing import Dict, Any, Optional

try:
    import psutil
except ImportError:
    psutil = None

try:
    import torch
except ImportError:
    torch = None


def _env_int(name: str, default: Optional[int] = None) -> Optional[int]:
    val = os.getenv(name)
    if val is None or str(val).strip() == "":
        return default
    try:
        return int(float(str(val).strip()))
    except Exception:
        return default


def _env_float(name: str, default: Optional[float] = None) -> Optional[float]:
    val = os.getenv(name)
    if val is None or str(val).strip() == "":
        return default
    try:
        return float(str(val).strip())
    except Exception:
        return default


def detect_cgroup_cpu_count() -> Optional[int]:
    """Detect Docker/Kubernetes CPU quota if running inside a container."""
    # cgroup v2
    try:
        if os.path.exists("/sys/fs/cgroup/cpu.max"):
            with open("/sys/fs/cgroup/cpu.max", "r", encoding="utf-8") as f:
                parts = f.read().strip().split()
                if len(parts) == 2 and parts[0] != "max":
                    quota = int(parts[0])
                    period = int(parts[1])
                    if quota > 0 and period > 0:
                        return max(1, int(math.ceil(quota / period)))
    except Exception:
        pass
    # cgroup v1
    try:
        q_file = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
        p_file = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
        if os.path.exists(q_file) and os.path.exists(p_file):
            with open(q_file, "r", encoding="utf-8") as fq:
                quota = int(fq.read().strip())
            with open(p_file, "r", encoding="utf-8") as fp:
                period = int(fp.read().strip())
            if quota > 0 and period > 0:
                return max(1, int(math.ceil(quota / period)))
    except Exception:
        pass
    return None


def detect_cgroup_memory_mb() -> Optional[int]:
    """Detect Docker/Kubernetes RAM limit in MB if running inside a container."""
    for path in ("/sys/fs/cgroup/memory.max", "/sys/fs/cgroup/memory/memory.limit_in_bytes"):
        try:
            if os.path.exists(path):
                with open(path, "r", encoding="utf-8") as f:
                    raw = f.read().strip()
                if raw and raw != "max":
                    val = int(raw)
                    # Ignore unlimited sentinel values (> 1 PB)
                    if 0 < val < (1 << 50):
                        return max(256, int(val // (1024 * 1024)))
        except Exception:
            pass
    return None


def get_system_resources(override_vcpu: Optional[int] = None, override_ram_mb: Optional[int] = None) -> Dict[str, Any]:
    """Return current hardware/container vCPU, total RAM, available RAM, and CPU load."""
    env_vcpu = override_vcpu or _env_int("AI_HOST_VCPU")
    cg_vcpu = detect_cgroup_cpu_count()
    os_vcpu = os.cpu_count() or 2
    vcpu = env_vcpu or cg_vcpu or os_vcpu
    vcpu = max(1, int(vcpu))

    env_ram = override_ram_mb or _env_int("AI_HOST_RAM_MB")
    cg_ram = detect_cgroup_memory_mb()
    host_ram_mb = 8192
    avail_ram_mb = 4096
    cpu_percent = 10.0
    process_rss_mb = 0.0

    if psutil is not None:
        try:
            vm = psutil.virtual_memory()
            host_ram_mb = int(vm.total // (1024 * 1024))
            avail_ram_mb = int(vm.available // (1024 * 1024))
        except Exception:
            pass
        try:
            cpu_percent = float(psutil.cpu_percent(interval=None))
        except Exception:
            pass
        try:
            process_rss_mb = round(psutil.Process(os.getpid()).memory_info().rss / (1024 * 1024), 1)
        except Exception:
            pass

    total_ram_mb = env_ram or (min(host_ram_mb, cg_ram) if cg_ram else host_ram_mb)
    total_ram_mb = max(512, int(total_ram_mb))
    avail_ram_mb = min(total_ram_mb, max(128, int(avail_ram_mb)))

    return {
        "vcpu": vcpu,
        "ram_total_mb": total_ram_mb,
        "ram_available_mb": avail_ram_mb,
        "cpu_percent": round(cpu_percent, 1),
        "process_rss_mb": process_rss_mb,
    }


def get_resource_profile(override_vcpu: Optional[int] = None, override_ram_mb: Optional[int] = None) -> Dict[str, Any]:
    """
    Compute the active AI Training Resource Profile.
    Re-reads environment variables on every call so ops can tune parameters dynamically.
    """
    sys_res = get_system_resources(override_vcpu=override_vcpu, override_ram_mb=override_ram_mb)
    vcpu = sys_res["vcpu"]
    ram_total_mb = sys_res["ram_total_mb"]
    ram_avail_mb = sys_res["ram_available_mb"]

    # Auto-detect concurrency: max(1, floor((vCPU - 1) / 2))
    # 2 vCPU -> 1, 4 vCPU -> 1, 6 vCPU -> 2, 8 vCPU -> 3, 16 vCPU -> 7
    auto_concurrent = max(1, (vcpu - 1) // 2)
    max_concurrent = _env_int("AI_TRAIN_MAX_CONCURRENT", auto_concurrent)
    max_concurrent = max(1, int(max_concurrent or 1))

    # Auto-detect threads per training job: reserve at least 1 vCPU for FastAPI inference
    auto_threads = max(1, (vcpu - 1) // max_concurrent)
    train_threads = _env_int("AI_TRAIN_THREADS", auto_threads)
    train_threads = max(1, int(train_threads or 1))

    # RAM budget for training (~38-40% of total RAM, e.g. ~3000 MB on 8 GB)
    auto_max_ram = max(512, int(round(ram_total_mb * 0.375, -2)))
    max_ram_mb = _env_int("AI_TRAIN_MAX_RAM_MB", auto_max_ram)
    max_ram_mb = max(256, int(max_ram_mb or auto_max_ram))

    # Minimum free RAM required before starting a job (~18% of total RAM, e.g. 1500 MB on 8 GB)
    auto_min_free = max(256, min(1500 if ram_total_mb <= 8192 else int(ram_total_mb * 0.18), int(ram_total_mb * 0.25)))
    min_free_ram_mb = _env_int("AI_TRAIN_MIN_FREE_RAM_MB", auto_min_free)
    min_free_ram_mb = max(128, int(min_free_ram_mb or auto_min_free))

    # Batch size scales with RAM budget (5,000 rows on 8 GB, up to 50,000 on 64 GB+)
    auto_batch_rows = max(2000, min(50000, int((max_ram_mb / 3000.0) * 5000)))
    batch_rows = _env_int("BATCH_ROWS", auto_batch_rows)
    batch_rows = max(500, int(batch_rows or 5000))

    # Tier thresholds (in rows)
    t1_max_rows = _env_int("AI_TRAIN_T1_MAX_ROWS", int(200_000 * max(1.0, max_ram_mb / 3000.0)))
    t2_max_rows = _env_int("AI_TRAIN_T2_MAX_ROWS", 50_000_000)
    t3_sample_rows = _env_int("AI_TRAIN_MAX_ROWS", int(min(5_000_000, max(250_000, batch_rows * 100))))

    auto_interval = max(10, int(_env_int("RETRAIN_AUTO_INTERVAL", 60) or 60))
    debounce_seconds = max(0, int(_env_int("RETRAIN_DEBOUNCE_SECONDS", 60) or 60))
    cpu_max_percent = float(_env_float("AI_TRAIN_MAX_CPU_PERCENT", 90.0) or 90.0)
    metric_tolerance = float(_env_float("AI_TRAIN_METRIC_TOLERANCE", 0.03) or 0.03)
    executor_mode = os.getenv("AI_TRAIN_EXECUTOR", "thread").strip().lower()
    if executor_mode not in ("thread", "process", "remote"):
        executor_mode = "thread"

    device = os.getenv("AI_TRAIN_DEVICE", "").strip().lower()
    if not device:
        device = "cuda" if (torch is not None and torch.cuda.is_available()) else "cpu"

    guard_ok = True
    guard_reason = None
    if ram_avail_mb < min_free_ram_mb:
        guard_ok = False
        guard_reason = f"Menunggu resource (RAM bebas {ram_avail_mb} MB < {min_free_ram_mb} MB)"
    elif sys_res["cpu_percent"] > cpu_max_percent:
        guard_ok = False
        guard_reason = f"Menunggu resource (Beban CPU {sys_res['cpu_percent']}% > {cpu_max_percent}%)"

    return {
        **sys_res,
        "max_concurrent": max_concurrent,
        "train_threads": train_threads,
        "max_ram_mb": max_ram_mb,
        "min_free_ram_mb": min_free_ram_mb,
        "batch_rows": batch_rows,
        "t1_max_rows": t1_max_rows,
        "t2_max_rows": t2_max_rows,
        "t3_sample_rows": t3_sample_rows,
        "auto_interval": auto_interval,
        "debounce_seconds": debounce_seconds,
        "cpu_max_percent": cpu_max_percent,
        "metric_tolerance": metric_tolerance,
        "executor": executor_mode,
        "device": device,
        "guard_ok": guard_ok,
        "guard_reason": guard_reason,
    }


def select_training_tier(row_count: int, est_bytes_per_row: int = 256, profile: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    """
    Select training tier based on dataset row count and available RAM budget:
    - T1 (Small / In-Memory): < ~200k rows and fits comfortably inside RAM budget
    - T2 (Medium / Out-of-Core Batch): 200k - 50M rows (Parquet batch streaming / incremental boosting / partial_fit)
    - T3 (Large / TB-Scale): > 50M rows or exceeds out-of-core budget (Stratified / minority-preserving reservoir sampling + incremental)
    """
    prof = profile or get_resource_profile()
    t1_max = prof["t1_max_rows"]
    t2_max = prof["t2_max_rows"]
    max_ram_bytes = prof["max_ram_mb"] * 1024 * 1024
    est_dataset_bytes = max(0, int(row_count)) * max(64, int(est_bytes_per_row))

    # Working memory multiplier (~4x raw row bytes for feature matrix + model trees)
    fits_in_ram_budget = (est_dataset_bytes * 4) <= max_ram_bytes

    if row_count <= t1_max and fits_in_ram_budget:
        tier = "T1"
        strategy = "in_memory_batch"
        target_rows = row_count
    elif row_count <= t2_max:
        tier = "T2"
        strategy = "out_of_core_incremental"
        target_rows = row_count
    else:
        tier = "T3"
        strategy = "stratified_reservoir_sampling"
        target_rows = min(row_count, prof["t3_sample_rows"])

    total_batches = max(1, int(math.ceil(max(1, target_rows) / float(prof["batch_rows"]))))
    return {
        "tier": tier,
        "strategy": strategy,
        "input_rows": int(row_count),
        "target_rows": int(target_rows),
        "batch_rows": int(prof["batch_rows"]),
        "total_batches": total_batches,
        "est_ram_mb": round(min(prof["max_ram_mb"], (target_rows * est_bytes_per_row * 3) / (1024 * 1024)), 2),
    }


def apply_thread_limits(threads: Optional[int] = None) -> int:
    """Apply CPU thread limits and low process priority so training never starves the FastAPI worker."""
    prof = get_resource_profile()
    n_threads = max(1, int(threads or prof["train_threads"]))
    for env_key in ("OMP_NUM_THREADS", "MKL_NUM_THREADS", "OPENBLAS_NUM_THREADS", "NUMEXPR_NUM_THREADS"):
        os.environ[env_key] = str(n_threads)
    if torch is not None:
        try:
            torch.set_num_threads(n_threads)
        except Exception:
            pass
    return n_threads


def lower_current_process_priority():
    """Set below-normal scheduling priority on the worker if supported by the OS."""
    if psutil is None:
        return
    try:
        p = psutil.Process(os.getpid())
        if os.name == "nt" and hasattr(psutil, "BELOW_NORMAL_PRIORITY_CLASS"):
            p.nice(psutil.BELOW_NORMAL_PRIORITY_CLASS)
        elif hasattr(os, "nice"):
            current = p.nice()
            if current < 10:
                p.nice(10)
    except Exception:
        pass
