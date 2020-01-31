pip install openapi2jsonschema

schema='https://raw.githubusercontent.com/kubernetes/kubernetes/v1.17.0/api/openapi-spec/swagger.json'

openapi2jsonschema -o "k8s-schema" --stand-alone "${schema}"
