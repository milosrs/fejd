#!/bin/sh
set -eu

if [ -n "${RESEND_API_KEY:-}" ]; then
    cat > /etc/smtpd/smtpd.conf <<EOF
listen on 0.0.0.0 port 25

table secrets { "resend" = "resend:${RESEND_API_KEY}" }

action "relay" relay host smtps://resend@smtp.resend.com:465 auth <secrets>

match from any for any action "relay"
EOF
else
    cat > /etc/smtpd/smtpd.conf <<EOF
listen on 0.0.0.0 port 25

table vusers { "@" = "mailuser" }

action "catchall" mda "/bin/cat >> /var/mail/catchall.eml" user mailuser virtual <vusers>

match from any for any action "catchall"
EOF
fi

exec smtpd -d
