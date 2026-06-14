FROM docker.elastic.co/elasticsearch/elasticsearch:8.15.5

ARG IK_VERSION=8.15.5

RUN elasticsearch-plugin install --batch \
    "https://release.infinilabs.com/analysis-ik/stable/elasticsearch-analysis-ik-${IK_VERSION}.zip"
