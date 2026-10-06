"""
AuditSphere Data Hub — Hardware Adaptor
Detects available memory & cores to dynamically adjust ingestion chunk size and concurrency.
"""
import os

def get_hardware_profile():
    profile = os.getenv("HARDWARE_PROFILE", "auto").lower()
    
    # Defaults for demo VPS (KVM 2: 2 vCPU, 8GB RAM)
    demo_config = {
        "profile": "demo",
        "chunk_size": int(os.getenv("INGEST_CHUNK_ROWS", 50000)),
        "max_parallel_tables": 1,
        "max_connections_per_source": 2,
        "sample_size_ai": 200000,
        "statement_timeout_ms": int(os.getenv("CLIENT_DB_STATEMENT_TIMEOUT", 60000))
    }
    
    # Defaults for production server (1TB+ dataset, 16+ vCPU, 64GB+ RAM)
    prod_config = {
        "profile": "prod",
        "chunk_size": int(os.getenv("INGEST_CHUNK_ROWS", 100000)),
        "max_parallel_tables": 3,
        "max_connections_per_source": 4,
        "sample_size_ai": 1000000,
        "statement_timeout_ms": int(os.getenv("CLIENT_DB_STATEMENT_TIMEOUT", 60000))
    }
    
    if profile == "demo":
        return demo_config
    elif profile == "prod":
        return prod_config
    
    # Auto detection based on system memory
    try:
        if os.path.exists("/proc/meminfo"):
            with open("/proc/meminfo", "r") as f:
                for line in f:
                    if line.startswith("MemTotal:"):
                        mem_kb = int(line.split()[1])
                        mem_mb = mem_kb // 1024
                        if mem_mb > 32768: # > 32GB RAM
                            return prod_config
                        else:
                            return demo_config
    except Exception:
        pass
        
    return demo_config
