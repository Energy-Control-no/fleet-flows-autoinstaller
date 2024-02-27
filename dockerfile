FROM golang:latest

WORKDIR /auto-installer

COPY . .

RUN go get

RUN go build -o installer .

ENV AIRTABLE_BASE_ID=appYWVOaoPhQB0nmA
ENV FILES_BRANCH=main
ENV AIRTABLE_API_KEY=
ENV GIT_SERVER=ssh://git@fleet-flows-git.lizzardsolutions.com
ENV FLOW_JS_BRANCH=main
ENV AIRTABLE_TABLE=Unipi
ENV MAX_RETRIES=5
ENV SLEEP_BETWEEN=5
ENV SCHEMA_FILE_PATH=

# VOLUME [ "/auto-installer/configs" ]

# CMD [ "/auto-installer/installer", "-k="pattYv2J5IWZdl7HF.29608e738acb1056754b8c3441b0bf0a25b9c900b8f9bd606526aa8e7ec185c5"", "-t="Unipi"", "-b="appYWVOaoPhQB0nmA"" ]
CMD ["/auto-installer/installer", "-k=pattYv2J5IWZdl7HF.29608e738acb1056754b8c3441b0bf0a25b9c900b8f9bd606526aa8e7ec185c5", "-t=Unipi", "-b=appYWVOaoPhQB0nmA"]
