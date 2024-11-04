FROM python:3.12.4-slim

WORKDIR /flask

COPY requirements.txt requirements.txt
COPY config.py config.py
COPY wsgi.py wsgi.py
COPY app app

COPY --chmod=755 entrypoint.sh /entrypoint.sh

RUN : \
  && python3 -m venv .venv \
  && . .venv/bin/activate \
  && pip install --upgrade pip \
  && pip install -r requirements.txt \
  && :
  
EXPOSE 8081

ENTRYPOINT ["/entrypoint.sh"]
CMD ["gunicorn", "-w 4", "-b 0.0.0.0:8081", "wsgi:app"]
