FROM python:3.10-slim

WORKDIR /app

# ------------------------------------------------------------------
# Set PyPI as the DEFAULT index for ALL pip invocations — including
# any subprocess pip spawns to install build backends (flit_core etc).
# This is the root fix for the flit_core / typing_extensions error.
# ------------------------------------------------------------------
ENV PIP_INDEX_URL=https://pypi.org/simple
ENV PIP_NO_CACHE_DIR=1

# ------------------------------------------------------------------
# Layer 1: upgrade pip and pre-install every transitive dep that
# torch needs, all from plain PyPI — no PyTorch index involved yet.
# ------------------------------------------------------------------
RUN pip install --upgrade pip setuptools wheel && \
    pip install \
        typing-extensions \
        jinja2 \
        filelock \
        sympy \
        networkx \
        fsspec \
        mpmath \
        pillow \
        numpy==1.26.4

# ------------------------------------------------------------------
# Layer 2: install torch CPU wheels via --find-links.
# --find-links points Docker at PyTorch's wheel mirror but does NOT
# override PIP_INDEX_URL, so any dependency subprocess still uses
# PyPI. --no-deps means torch itself won't trigger more resolution.
# ------------------------------------------------------------------
RUN pip install \
        torch==2.2.2+cpu \
        torchvision==0.17.2+cpu \
        torchaudio==2.2.2+cpu \
        --no-deps \
        --find-links https://download.pytorch.org/whl/cpu/torch_stable.html

# ------------------------------------------------------------------
# Layer 3: app dependencies
# ------------------------------------------------------------------
COPY requirements.txt .
RUN pip install --default-timeout=1000 --retries 10 -r requirements.txt

COPY . .

# Backup pre-baked models to avoid them being shadowed by empty mounted volumes
RUN mkdir -p /app/pretrained_models && cp -r models/* /app/pretrained_models/ || true

# Setup entrypoint script
RUN chmod +x entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]