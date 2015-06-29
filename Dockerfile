FROM progrium/busybox
RUN opkg-install ca-certificates
RUN cat /etc/ssl/certs/*.crt > /etc/ssl/certs/ca-certificates.crt && \
    sed -i -r '/^#.+/d' /etc/ssl/certs/ca-certificates.crt
ADD auth /bin/auth
CMD /bin/auth
