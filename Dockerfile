ARG SOURCE

FROM python:3.11-alpine

ARG SOURCE=$SOURCE

WORKDIR /airbyte/integration_code/

COPY ./connectors/$SOURCE /airbyte/integration_code

RUN pip install .

ENV AIRBYTE_ENTRYPOINT "python /airbyte/integration_code/main.py"
ENTRYPOINT ["python", "/airbyte/integration_code/main.py"]