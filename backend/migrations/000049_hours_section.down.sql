DELETE FROM translations WHERE key LIKE 'sections.hours%';

ALTER TABLE sections DROP CONSTRAINT sections_type_check;
ALTER TABLE sections ADD CONSTRAINT sections_type_check
    CHECK (type IN ('hero', 'about', 'gallery', 'contact'));
