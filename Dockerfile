FROM golang:1.16 as build

WORKDIR /go/flashpaper
COPY . .
RUN go mod vendor && CGO_ENABLED=0 go build -o /flashpaper .

FROM scratch
COPY --from=build /flashpaper /flashpaper
COPY theme /theme
ENTRYPOINT ["/flashpaper"]
