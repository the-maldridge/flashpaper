FROM golang:1.17 as build

WORKDIR /go/flashpaper
COPY . .
RUN go mod vendor && CGO_ENABLED=0 go build -o /flashpaper .

FROM scratch
COPY --from=build /flashpaper /flashpaper
COPY theme /theme
ENTRYPOINT ["/flashpaper"]
EXPOSE 8080/tcp
