CREATE TABLE common_reference_master (
    id BIGSERIAL PRIMARY KEY,
    entity VARCHAR(255) NOT NULL,
    entity_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tax status master table
CREATE TABLE tax_status_master (
    id BIGSERIAL PRIMARY KEY,
    tax_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- UCC table
CREATE TABLE ucc (
    ucc_id BIGSERIAL PRIMARY KEY,
    client_code VARCHAR(255),
    member_code VARCHAR(255),
    tax_status BIGINT REFERENCES tax_status_master(id),
    holding_nature BIGINT REFERENCES common_reference_master(id),
    ucc_status BIGINT REFERENCES common_reference_master(id),
    is_client_physical BOOLEAN NOT NULL DEFAULT FALSE,
    is_client_demat BOOLEAN NOT NULL DEFAULT FALSE,
    parent_client_code VARCHAR(255),
    deleted_at TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Holder table
CREATE TABLE holder (
    id BIGSERIAL PRIMARY KEY,
    ref_id BIGINT NOT NULL REFERENCES ucc(ucc_id) ON DELETE CASCADE,
    holder_rank BIGINT REFERENCES common_reference_master(id),
    deleted_at TIMESTAMPTZ,
    deleted_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Person detail table
CREATE TABLE person_detail (
    id BIGSERIAL PRIMARY KEY,
    ref_id BIGINT NOT NULL REFERENCES holder(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    first_name VARCHAR(255),
    middle_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Non-individual detail table
CREATE TABLE non_individual_detail (
    id BIGSERIAL PRIMARY KEY,
    ref_id BIGINT NOT NULL REFERENCES holder(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    org_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optional indexes for performance
CREATE INDEX idx_ucc_client_code ON ucc(client_code);
CREATE INDEX idx_ucc_member_code ON ucc(member_code);
CREATE INDEX idx_holder_ref_id ON holder(ref_id);
CREATE INDEX idx_person_detail_ref_id ON person_detail(ref_id);
CREATE INDEX idx_non_individual_detail_ref_id ON non_individual_detail(ref_id);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_non_individual_detail_ref_id;
DROP INDEX IF EXISTS idx_person_detail_ref_id;
DROP INDEX IF EXISTS idx_holder_ref_id;
DROP INDEX IF EXISTS idx_ucc_member_code;
DROP INDEX IF EXISTS idx_ucc_client_code;

DROP TABLE IF EXISTS non_individual_detail CASCADE;
DROP TABLE IF EXISTS person_detail CASCADE;
DROP TABLE IF EXISTS holder CASCADE;
DROP TABLE IF EXISTS ucc CASCADE;
DROP TABLE IF EXISTS tax_status_master CASCADE;
DROP TABLE IF EXISTS common_reference_master CASCADE;
