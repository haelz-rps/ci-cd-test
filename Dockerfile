FROM python:3.11-alpine

ARG SOURCE

COPY ./connectors/source-sage /airbyte/integration_code

WORKDIR /airbyte/integration_code/

RUN chmod -R 755 /airbyte/integration_code
RUN pip install .

ENV AIRBYTE_ENTRYPOINT "python /airbyte/integration_code/main.py"
ENTRYPOINT ["python", "/airbyte/integration_code/main.py"]
