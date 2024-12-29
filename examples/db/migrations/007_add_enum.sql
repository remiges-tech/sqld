-- Create enum type for employee status
CREATE TYPE employee_status AS ENUM ('active', 'on_leave', 'terminated');

-- Add status column to employees table
ALTER TABLE employees ADD COLUMN status employee_status DEFAULT 'active';

-- Update some existing records for testing
UPDATE employees SET status = 'on_leave' WHERE id % 3 = 0;
UPDATE employees SET status = 'terminated' WHERE id % 5 = 0;

---- create above / drop below ----

ALTER TABLE employees DROP COLUMN status;
DROP TYPE employee_status;
