# SQLD Architecture

## Overview

SQLD provides type-safe dynamic SQL query execution with two subsystems:

1. **Structured Query System**: For executing queries using a JSON-based request format with field validation
2. **Raw Query System**: For executing custom SQL with named parameters and type checking

## Core Components

### 1. Registry
- Stores model metadata (table names and fields)
- Maps between database fields and JSON fields
- Validates field names and types at runtime

### 2. Validator
- Validates query fields against model metadata
- Checks pagination parameters
- Ensures field names exist in models
- Validates WHERE clause fields

### 3. Query Execution
- Executes queries with proper type safety
- Supports both sql.DB and pgx.Conn
- Maps results to correct field types
- Handles pagination metadata

## Features

### Structured Queries
- Select specific fields with validation
- WHERE clause with field checking
- ORDER BY with field validation
- Pagination options:
  - Page-based (page/page_size)
  - Direct (limit/offset)

### Raw Queries
- Custom SQL with named parameters
- Runtime type validation
- SQL syntax checking
- Field filtering in results

## Dependencies

- squirrel: For building SQL queries
- scany: For scanning query results
- pgx: For PostgreSQL support
