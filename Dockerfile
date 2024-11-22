FROM python:3.11.10-alpine3.20


ARG SOURCE

COPY ./connectors/$SOURCE /airbyte/integration_code

WORKDIR /airbyte/integration_code/

RUN poetry install

ENV AIRBYTE_ENTRYPOINT "python /airbyte/integration_code/main.py"
ENTRYPOINT ["python", "/airbyte/integration_code/main.py"]
