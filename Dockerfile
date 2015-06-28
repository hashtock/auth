FROM debian:jessie
RUN apt-get update; apt-get install -y -qq ca-certificates
ADD auth /bin/auth
CMD /bin/auth
