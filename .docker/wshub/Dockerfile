FROM golang:1.25-bookworm

WORKDIR /app

RUN addgroup --gid=1000 app
RUN addgroup --gid=1002 developer
RUN adduser --disabled-password --quiet --uid=1000 --gid=1000 --no-create-home --gecos GECOS app
RUN mkdir -p /home/app
RUN chown 1000:1000 /home/app

ENV GOCACHE=/go-cache

#CMD ["sleep", "infinity"]

CMD ["go", "run", "./cmd/wshub"]
