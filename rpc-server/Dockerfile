FROM golang:1.20
WORKDIR /app
COPY . .
RUN sh ./build.sh
EXPOSE 8888
CMD ["./output/bootstrap.sh"]
