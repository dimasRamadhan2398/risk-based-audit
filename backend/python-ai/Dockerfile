FROM python:3.10-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir torch==2.2.2+cpu torchvision==0.17.2+cpu torchaudio==2.2.2+cpu --extra-index-url https://download.pytorch.org/whl/cpu
RUN pip install --default-timeout=1000 --retries 10 --no-cache-dir -i https://mirrors.aliyun.com/pypi/simple/ -r requirements.txt

COPY . .

# Run the training script during the build so models are ready
# Ignore errors if IndoBERT fails to download
RUN python train.py || true

# Backup pre-baked models to avoid them being shadowed by empty mounted volumes
RUN mkdir -p /app/pretrained_models && cp -r models/* /app/pretrained_models/ || true

# Setup entrypoint script
RUN chmod +x entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
