CREATE TABLE family_relation_translations(
    relation_type_id UUID NOT NULL REFERENCES family_relation_types(id) ON DELETE CASCADE,
    locale VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (relation_type_id, locale),

    CONSTRAINT family_relation_translations_locale_not_blank
        CHECK (length(trim(locale)) > 0),

    CONSTRAINT family_relation_translations_name_not_blank
        CHECK (length(trim(name)) > 0)
);