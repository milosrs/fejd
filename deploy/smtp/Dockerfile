FROM alpine:3.20

RUN apk add --no-cache opensmtpd \
    && adduser -D mailuser \
    && mkdir -p /var/mail \
    && chown mailuser:mailuser /var/mail

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 25

ENTRYPOINT ["/entrypoint.sh"]
