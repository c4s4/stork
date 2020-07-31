FROM scratch

ADD build/stork /

ENTRYPOINT ["/stork"]
