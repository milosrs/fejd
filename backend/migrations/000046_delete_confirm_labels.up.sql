-- Delete confirmation labels (generic Yes/No prompt + section delete title).

INSERT INTO translations (key, locale, value) VALUES
    ('common.yes',            'en', 'Yes'),
    ('common.yes',            'rs', 'Da'),
    ('common.no',             'en', 'No'),
    ('common.no',             'rs', 'Ne'),
    ('common.deleteConfirm',  'en', 'This will delete {name}. Are you sure you want to proceed?'),
    ('common.deleteConfirm',  'rs', 'Ovo će obrisati {name}. Da li ste sigurni da želite da nastavite?'),
    ('landing.deleteSection', 'en', 'Delete section'),
    ('landing.deleteSection', 'rs', 'Obriši sekciju');
