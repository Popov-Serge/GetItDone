CREATE TABLE member_aliases (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                family_id UUID NOT NULL
                                    REFERENCES families(id)
                                        ON DELETE CASCADE,

                                family_member_id UUID NOT NULL
                                    REFERENCES family_members(id)
                                        ON DELETE CASCADE,

                                alias VARCHAR(100) NOT NULL,

                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                CONSTRAINT member_aliases_alias_not_blank
                                    CHECK (length(trim(alias)) > 0),

                                UNIQUE (family_member_id, alias)
);

CREATE UNIQUE INDEX ux_member_aliases_family_alias
    ON member_aliases (family_id, lower(alias));