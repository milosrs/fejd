-- Store the no-show grace period in minutes so owners can configure an
-- HH:MM duration (e.g. 00:35 or 01:22). Existing values are converted from
-- hours to minutes.
ALTER TABLE businesses ADD COLUMN no_show_after_minutes INTEGER NOT NULL DEFAULT 120;
UPDATE businesses SET no_show_after_minutes = no_show_after_hours * 60;
ALTER TABLE businesses DROP COLUMN no_show_after_hours;

-- No-show grace is now entered as HH:MM and is optional.
UPDATE translations SET value = 'No-show grace (optional)' WHERE key = 'policy.noShowGrace' AND locale = 'en';
UPDATE translations SET value = 'Grey period za nepojavljivanje (opciono)' WHERE key = 'policy.noShowGrace' AND locale = 'rs';
UPDATE translations SET value = 'No-show grace must be a valid time (HH:MM).' WHERE key = 'policy.noShowInvalid' AND locale = 'en';
UPDATE translations SET value = 'Grey period za nepojavljivanje mora biti važeće vreme (HH:MM).' WHERE key = 'policy.noShowInvalid' AND locale = 'rs';
