-- Revert to a whole-hours no-show grace period, truncating any minute
-- remainder (integer division).
ALTER TABLE businesses ADD COLUMN no_show_after_hours INTEGER NOT NULL DEFAULT 2;
UPDATE businesses SET no_show_after_hours = no_show_after_minutes / 60;
ALTER TABLE businesses DROP COLUMN no_show_after_minutes;

UPDATE translations SET value = 'No-show grace (hours after)' WHERE key = 'policy.noShowGrace' AND locale = 'en';
UPDATE translations SET value = 'Grey period za nepojavljivanje (sati posle)' WHERE key = 'policy.noShowGrace' AND locale = 'rs';
UPDATE translations SET value = 'No-show grace must be zero or greater.' WHERE key = 'policy.noShowInvalid' AND locale = 'en';
UPDATE translations SET value = 'Grey period za nepojavljivanje mora biti nula ili veći.' WHERE key = 'policy.noShowInvalid' AND locale = 'rs';
