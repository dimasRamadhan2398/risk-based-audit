FROM jupyter/scipy-notebook:python-3.10

USER root

# System dependencies for psycopg2
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libpq-dev && \
    rm -rf /var/lib/apt/lists/*

# Copy requirements and install
COPY caatt-engine/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt

# Copy notebooks
COPY notebooks/ /home/jovyan/notebooks/

# Ensure correct ownership
RUN chown -R jovyan:users /home/jovyan/notebooks

USER jovyan

WORKDIR /home/jovyan

EXPOSE 8888

CMD ["sh", "-c", "start-notebook.sh --NotebookApp.token=\"${JUPYTER_TOKEN}\" --NotebookApp.allow_origin='*' --NotebookApp.base_url=/"]

