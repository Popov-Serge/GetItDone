CREATE TABLE family_relation_types(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT family_relation_types_code_not_blank
        CHECK (length(trim(code)) > 0),

    CONSTRAINT family_relation_types_sort_order_positive
        CHECK (sort_order >= 0)
);

INSERT INTO family_relation_types (code, sort_order)
VALUES
    ('father', 1),
    ('mother', 2),
    ('son', 3),
    ('daughter', 4),
    ('grandfather', 5),
    ('grandmother', 6),
    ('other', 7);