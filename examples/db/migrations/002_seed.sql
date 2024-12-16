-- Seed data for employees table
INSERT INTO employees (
    first_name, last_name, email, phone, 
    hire_date, salary, department, position, is_active
) VALUES 
    ('John', 'Smith', 'john.smith@company.com', '555-0101', 
    '2020-01-15', 75000.00, 'Engineering', 'Senior Developer', true),
    
    ('Sarah', 'Johnson', 'sarah.j@company.com', '555-0102', 
    '2021-03-20', 85000.00, 'Engineering', 'Tech Lead', true),
    
    ('Michael', 'Brown', 'michael.b@company.com', '555-0103', 
    '2019-11-01', 65000.00, 'Marketing', 'Marketing Manager', true),
    
    ('Emily', 'Davis', 'emily.d@company.com', '555-0104', 
    '2022-01-10', 55000.00, 'Sales', 'Sales Representative', true),
    
    ('David', 'Wilson', 'david.w@company.com', '555-0105', 
    '2020-06-15', 95000.00, 'Engineering', 'Principal Engineer', true),
    
    ('Lisa', 'Anderson', 'lisa.a@company.com', '555-0106', 
    '2021-09-01', 70000.00, 'HR', 'HR Manager', true),
    
    ('James', 'Taylor', 'james.t@company.com', '555-0107', 
    '2022-02-28', 60000.00, 'Marketing', 'Content Specialist', true),
    
    ('Maria', 'Garcia', 'maria.g@company.com', '555-0108', 
    '2020-08-15', 72000.00, 'Sales', 'Senior Sales Rep', true),
    
    ('Robert', 'Martinez', 'robert.m@company.com', '555-0109', 
    '2021-07-01', 68000.00, 'Engineering', 'Software Developer', true),
    
    ('Jennifer', 'Lee', 'jennifer.l@company.com', '555-0110', 
    '2019-12-01', 82000.00, 'HR', 'HR Director', true);

---- create above / drop below ----

-- Remove seed data
DELETE FROM employees WHERE email IN (
    'john.smith@company.com',
    'sarah.j@company.com',
    'michael.b@company.com',
    'emily.d@company.com',
    'david.w@company.com',
    'lisa.a@company.com',
    'james.t@company.com',
    'maria.g@company.com',
    'robert.m@company.com',
    'jennifer.l@company.com'
);
