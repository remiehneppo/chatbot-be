FROM semitechnologies/weaviate:1.28.2

ENV QUERY_DEFAULTS_LIMIT=25 \
    AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true \
    PERSISTENCE_DATA_PATH=/var/lib/weaviate \
    DEFAULT_VECTORIZER_MODULE=text2vec-ollama \
    ENABLE_MODULES=text2vec-ollama,generative-ollama \
    CLUSTER_HOSTNAME=node1

EXPOSE 8080 50051

VOLUME /var/lib/weaviate

COPY ./weaviate_data /var/lib/weaviate

CMD ["--host", "0.0.0.0", "--port", "8080", "--scheme", "http"]