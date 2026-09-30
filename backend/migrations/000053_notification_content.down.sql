ALTER TABLE businesses
    DROP COLUMN IF EXISTS reminder_body,
    DROP COLUMN IF EXISTS reminder_title,
    DROP COLUMN IF EXISTS staff_notifications_enabled;

DELETE FROM translations WHERE key IN
    ('policy.staffNotifications', 'policy.staffNotificationsHelp',
     'policy.reminderTitle', 'policy.reminderBody', 'policy.reminderBodyHelp');
