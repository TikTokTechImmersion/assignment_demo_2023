FROM golang:1.18
WORKDIR /app
COPY . .
RUN sh ./build.sh
EXPOSE 8888
CMD ["./output/bootstrap.sh"]
