CREATE TABLE family_members (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                family_id UUID NOT NULL
                                    REFERENCES families(id)
                                        ON DELETE CASCADE,

                                user_id UUID
                                               REFERENCES users(id)
                                                   ON DELETE SET NULL,

                                name VARCHAR(100) NOT NULL,

                                relation_type_id UUID NOT NULL
                                    REFERENCES family_relation_types(id)
                                        ON DELETE RESTRICT,

                                role VARCHAR(20) NOT NULL DEFAULT 'member',

                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                CONSTRAINT family_members_name_not_blank
                                    CHECK (length(trim(name)) > 0),

                                CONSTRAINT family_members_role_check
                                    CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE UNIQUE INDEX ux_family_members_family_user
    ON family_members (family_id, user_id)
    WHERE user_id IS NOT NULL;