CREATE TABLE family_invitations (
                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                    family_id UUID NOT NULL
                                        REFERENCES families(id)
                                            ON DELETE CASCADE,

                                    family_member_id UUID NOT NULL
                                        REFERENCES family_members(id)
                                            ON DELETE CASCADE,

                                    token_hash CHAR(64) NOT NULL UNIQUE,

                                    expires_at TIMESTAMPTZ NOT NULL,

                                    accepted_at TIMESTAMPTZ,

                                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                    CONSTRAINT family_invitations_expiration_check
                                        CHECK (expires_at > created_at),

                                    CONSTRAINT family_invitations_accepted_check
                                        CHECK (
                                            accepted_at IS NULL
                                                OR accepted_at >= created_at
                                            )
);

CREATE INDEX ix_family_invitations_member
    ON family_invitations (family_member_id);

CREATE INDEX ix_family_invitations_family
    ON family_invitations (family_id);

CREATE UNIQUE INDEX ux_family_invitations_active_member
    ON family_invitations (family_member_id)
    WHERE accepted_at IS NULL;