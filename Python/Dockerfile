FROM python:alpine 

COPY app/ /app
WORKDIR /app

RUN pip install -r requirements.txt

ENV FLASK_APP application.py

EXPOSE 5000

CMD python3 application.py

