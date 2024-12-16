-- +seed Up

-- Seed common_reference_master with references for holder rank, holding nature, and UCC status
INSERT INTO common_reference_master (entity, entity_key, created_at, updated_at)
VALUES 
    ('UCC_HOLDER_RANK', 'FIRST', NOW(), NOW()),           -- id=1
    ('UCC_HOLDING_TYPE', 'INDIVIDUAL', NOW(), NOW()),     -- id=2
    ('UCC_ACC_STATUS', 'ACTIVE', NOW(), NOW());           -- id=3

-- Seed tax_status_master
INSERT INTO tax_status_master (tax_name, created_at, updated_at)
VALUES ('RESIDENT', NOW(), NOW()); -- id=1

-- Seed UCC (Unified Client Code)
INSERT INTO ucc (
    client_code, 
    member_code, 
    tax_status, 
    holding_nature, 
    ucc_status, 
    is_client_physical, 
    is_client_demat, 
    parent_client_code, 
    created_at, 
    updated_at
)
VALUES (
    'CL001',    -- client_code
    'MBR001',   -- member_code
    1,          -- tax_status (RESIDENT)
    2,          -- holding_nature (INDIVIDUAL)
    3,          -- ucc_status (ACTIVE)
    TRUE,       -- is_client_physical
    FALSE,      -- is_client_demat
    'PRNT001',  -- parent_client_code
    NOW(),
    NOW()
);

-- Seed holder for the above UCC
-- ref_id references ucc(ucc_id), holder_rank references common_reference_master
INSERT INTO holder (
    ref_id, 
    holder_rank, 
    created_at, 
    updated_at
)
VALUES (
    1,  -- refers to ucc_id=1 from the inserted UCC row
    1,  -- corresponds to UCC_HOLDER_RANK with entity_key='FIRST' (id=1)
    NOW(), 
    NOW()
);

-- Seed person_detail
-- ref_id references holder(id). The first holder inserted got id=1.
INSERT INTO person_detail (
    ref_id, 
    entity_type, 
    first_name, 
    middle_name, 
    last_name, 
    created_at, 
    updated_at
)
VALUES (
    1,          -- holder.id = 1
    'HOLDER', 
    'John', 
    'A', 
    'Doe', 
    NOW(),
    NOW()
);

-- Optionally seed another scenario with non-individual detail
-- This is just to show variety; it's not required.
INSERT INTO ucc (
    client_code, member_code, tax_status, holding_nature, ucc_status, is_client_physical, is_client_demat, parent_client_code, created_at, updated_at
) VALUES (
    'CL002', 'MBR002', 1, 2, 3, FALSE, TRUE, 'PRNT002', NOW(), NOW()
);

INSERT INTO holder (ref_id, holder_rank, created_at, updated_at)
VALUES (2, 1, NOW(), NOW()); -- holder for ucc #2

INSERT INTO non_individual_detail (ref_id, entity_type, org_name, created_at, updated_at)
VALUES (2, 'HOLDER', 'Acme Corp', NOW(), NOW());

---- create above / drop below ----

-- Remove seeded data in reverse order
DELETE FROM non_individual_detail WHERE org_name = 'Acme Corp';
DELETE FROM person_detail WHERE first_name = 'John' AND last_name = 'Doe';
DELETE FROM holder WHERE ref_id IN (1, 2);
DELETE FROM ucc WHERE client_code IN ('CL001', 'CL002');
DELETE FROM tax_status_master WHERE tax_name = 'RESIDENT';
DELETE FROM common_reference_master WHERE entity IN ('UCC_HOLDER_RANK', 'UCC_HOLDING_TYPE', 'UCC_ACC_STATUS');
