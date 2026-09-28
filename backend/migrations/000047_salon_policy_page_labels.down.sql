DELETE FROM translations WHERE key IN (
    'policy.appointments',
    'policy.workingHoursHelp',
    'policy.workingHoursNonWorking',
    'policy.nonWorkingDaysPolicyHelp',
    'policy.notAuthorized'
);
