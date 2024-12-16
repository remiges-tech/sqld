INSERT INTO accounts (account_number, account_name, account_type, balance, currency, status, owner_id)
VALUES
    ('1234567890', 'John Doe', 'checking', 1000.00, 'USD', 'active', 1),
    ('9876543210', 'Jane Smith', 'savings', 5000.00, 'USD', 'active', 2),
    ('5555555555', 'Bob Johnson', 'checking', 2000.00, 'USD', 'active', 3);

-- Insert test accounts
INSERT INTO accounts (account_number, account_name, account_type, balance, currency, status, owner_id)
VALUES 
    ('SAV001', 'High Interest Savings', 'savings', 5000.00, 'USD', 'active', 1),
    ('SAV002', 'Emergency Fund', 'savings', 2500.00, 'USD', 'active', 1),
    ('CHK001', 'Daily Expenses', 'checking', 1000.00, 'USD', 'active', 2),
    ('SAV003', 'Vacation Fund', 'savings', 1500.00, 'USD', 'active', 2),
    ('SAV004', 'Retirement Fund', 'savings', 10000.00, 'USD', 'active', 3),
    ('CHK002', 'Business Account', 'checking', 7500.00, 'USD', 'inactive', 3);

---- create above / drop below ----

SELECT (1);