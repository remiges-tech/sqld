-- name: GetEmployee :one
SELECT * FROM employees WHERE id = $1;

-- name: UCCList :many
SELECT
    u.ucc_id,
    u.client_code,
    u.member_code,
    ts.tax_name AS tax_status,
    crm.entity_key AS holding_nature,
    ucc_status.entity_key AS ucc_status,
    u.is_client_physical,
    u.is_client_demat,
    u.parent_client_code,
    COALESCE(
        CASE
            WHEN pd.first_name IS NOT NULL THEN CONCAT(pd.first_name, ' ', pd.middle_name, ' ', pd.last_name)
            WHEN nid.org_name IS NOT NULL THEN nid.org_name
            ELSE ''
        END, 
    '') AS primary_holder_name
FROM
    ucc u
LEFT JOIN
    holder h ON u.ucc_id = h.ref_id 
    AND h.holder_rank = (
        SELECT id 
        FROM common_reference_master 
        WHERE entity = 'UCC_HOLDER_RANK' 
          AND entity_key = 'FIRST'
    )
    AND h.deleted_at IS NULL 
    AND h.deleted_by IS NULL
LEFT JOIN
    person_detail pd ON pd.ref_id = h.id AND h.ref_id = u.ucc_id AND pd.entity_type = 'HOLDER'
LEFT JOIN 
    non_individual_detail nid ON nid.ref_id = h.id AND nid.entity_type = 'HOLDER'
LEFT JOIN
    common_reference_master AS crm ON u.holding_nature = crm.id AND crm.entity = 'UCC_HOLDING_TYPE'
LEFT JOIN
    common_reference_master AS ucc_status ON u.ucc_status = ucc_status.id AND ucc_status.entity = 'UCC_ACC_STATUS'
LEFT JOIN
    tax_status_master AS ts ON u.tax_status = ts.id
WHERE
    -- For integer params, NULL means no filter:
    (sqlc.narg('ucc_status')::bigint IS NULL OR u.ucc_status = sqlc.narg('ucc_status')::bigint) AND
    (sqlc.narg('tax_status')::bigint IS NULL OR u.tax_status = sqlc.narg('tax_status')::bigint) AND
    (sqlc.narg('holding_nature')::bigint IS NULL OR u.holding_nature = sqlc.narg('holding_nature')::bigint) AND

    -- For boolean params, NULL means no filter:
    (sqlc.narg('is_client_physical')::boolean IS NULL OR u.is_client_physical = sqlc.narg('is_client_physical')::boolean) AND
    (sqlc.narg('is_client_demat')::boolean IS NULL OR u.is_client_demat = sqlc.narg('is_client_demat')::boolean) AND
    NOT (u.is_client_physical = false AND u.is_client_demat = false) AND

    u.deleted_at IS NULL AND u.deleted_by IS NULL AND

    -- For text params, NULL or '' means no filter:
    (sqlc.narg('client_code')::text IS NULL OR sqlc.narg('client_code')::text = '' OR u.client_code ILIKE sqlc.narg('client_code')::text || '%') AND
    (sqlc.narg('member_code')::text IS NULL OR sqlc.narg('member_code')::text = '' OR u.member_code ILIKE sqlc.narg('member_code')::text || '%') AND
    (sqlc.narg('parent_client_code')::text IS NULL OR sqlc.narg('parent_client_code')::text = '' OR u.parent_client_code ILIKE sqlc.narg('parent_client_code')::text || '%') AND
    (
        sqlc.narg('search')::text IS NULL
        OR sqlc.narg('search')::text = ''
        OR (
            CONCAT_WS(' ', 
                u.client_code,  
                u.member_code,
                u.parent_client_code,
                ts.tax_name, 
                crm.entity_key,
                COALESCE(
                    pd.first_name || ' ' || pd.middle_name || ' ' || pd.last_name,
                    nid.org_name
                ),
                ucc_status.entity_key,
                CASE WHEN u.is_client_physical THEN 'true' ELSE 'false' END,
                CASE WHEN u.is_client_demat THEN 'true' ELSE 'false' END,
                COALESCE(pd.first_name, nid.org_name)
            ) ILIKE '%' || sqlc.narg('search')::text || '%'
        )
    )
ORDER BY
    CASE 
        WHEN sqlc.narg('sort_by') = 'ucc_id' AND upper(sqlc.narg('sort_order')) = 'A' THEN u.ucc_id
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_by') = 'ucc_id' AND upper(sqlc.narg('sort_order')) = 'D' THEN u.ucc_id
    END DESC,
    CASE
        WHEN sqlc.narg('sort_by') = 'client_code' AND upper(sqlc.narg('sort_order')) = 'A' THEN u.client_code
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_by') = 'client_code' AND upper(sqlc.narg('sort_order')) = 'D' THEN u.client_code
    END DESC,
    CASE
        WHEN sqlc.narg('sort_by') = 'member_code' AND upper(sqlc.narg('sort_order')) = 'A' THEN u.member_code
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_by') = 'member_code' AND upper(sqlc.narg('sort_order')) = 'D' THEN u.member_code
    END DESC,
    CASE
        WHEN sqlc.narg('sort_by') = 'tax_status' AND upper(sqlc.narg('sort_order')) = 'A' THEN ts.tax_name
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_by') = 'tax_status' AND upper(sqlc.narg('sort_order')) = 'D' THEN ts.tax_name
    END DESC,
    CASE
        WHEN sqlc.narg('sort_by') = 'primary_holder_name' AND upper(sqlc.narg('sort_order')) = 'A' THEN COALESCE(pd.first_name, nid.org_name)
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_by') = 'primary_holder_name' AND upper(sqlc.narg('sort_order')) = 'D' THEN COALESCE(pd.first_name, nid.org_name)
    END DESC
OFFSET sqlc.narg('offset')::int
LIMIT sqlc.narg('limit')::int;

-- name: GetEmployeesAdvanced :many
SELECT 
    id, 
    first_name, 
    last_name, 
    department, 
    salary
FROM employees
WHERE 
    department = sqlc.arg('department')
    AND salary >= sqlc.arg('min_salary')
    AND salary <= sqlc.arg('max_salary')
ORDER BY salary DESC;

-- name: GetEmployeesWithAccounts :many
SELECT 
    e.first_name as employee_name,
    e.department as dept,
    COALESCE(COUNT(a.id), 0)::bigint as account_count,
    COALESCE(SUM(a.balance), 0)::numeric as total_balance
FROM employees e
LEFT JOIN accounts a ON a.owner_id = e.id
WHERE e.department = sqlc.arg('department')
AND e.salary >= sqlc.arg('min_salary')
AND e.salary <= sqlc.arg('max_salary')
GROUP BY e.first_name, e.department
ORDER BY total_balance DESC;
