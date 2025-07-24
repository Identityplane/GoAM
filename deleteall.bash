#!/bin/bash

# Database file path
DB_FILE="cmd/goiam.db"

# Function to check if database file exists
check_db() {
    if [ ! -f "$DB_FILE" ]; then
        echo "Error: Database file $DB_FILE not found!"
        exit 1
    fi
}

# Function to set admin entitlements for a user
set_admin() {
    local username="$1"
    
    if [ -z "$username" ]; then
        echo "Error: Username is required!"
        echo "Usage: $0 set_admin <username>"
        exit 1
    fi
    
    check_db
    
    # Check if user exists in users table with realm="internal" and tenant="internal"
    local user_exists=$(sqlite3 "$DB_FILE" "SELECT COUNT(*) FROM users WHERE username = '$username' AND realm = 'internal' AND tenant = 'internal';")
    
    if [ "$user_exists" -eq 0 ]; then
        echo "Error: User '$username' not found in users table with realm='internal' and tenant='internal'!"
        exit 1
    fi
    
    # Update user entitlements to admin level with realm="internal" and tenant="internal"
    sqlite3 "$DB_FILE" "UPDATE users SET entitlements = '[\"*/*/*\"]' WHERE username = '$username' AND realm = 'internal' AND tenant = 'internal';"
    
    echo "Successfully set admin entitlements for user '$username'"
}

# Function to list all tables
list_tables() {
    echo "Available tables:"
    sqlite3 "$DB_FILE" ".tables"
}

# Function to delete all records from a table
delete_table() {
    local table_name="$1"
    
    # Check if table exists by checking if it's in the list of tables
    local tables=$(sqlite3 "$DB_FILE" ".tables")
    local table_exists=false
    
    for table in $tables; do
        if [ "$table" = "$table_name" ]; then
            table_exists=true
            break
        fi
    done
    
    if [ "$table_exists" = false ]; then
        echo "Error: Table '$table_name' does not exist!"
        echo ""
        list_tables
        exit 1
    fi
    
    # Get count before deletion
    local count=$(sqlite3 "$DB_FILE" "SELECT COUNT(*) FROM $table_name;")
    
    if [ "$count" -eq 0 ]; then
        echo "Table '$table_name' is already empty."
        return
    fi
    
    # Confirm deletion
    echo "About to delete $count records from table '$table_name'"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Delete all records
        sqlite3 "$DB_FILE" "DELETE FROM $table_name;"
        echo "Deleted $count records from table '$table_name'"
    else
        echo "Operation cancelled."
    fi
}

# Main script logic
main() {
    check_db
    
    if [ $# -eq 0 ]; then
        # No parameters provided, list all tables
        list_tables
    elif [ "$1" = "set_admin" ]; then
        # Handle set_admin command
        if [ $# -lt 2 ]; then
            echo "Error: Username required for set_admin command!"
            echo "Usage: $0 set_admin <username>"
            exit 1
        fi
        set_admin "$2"
    else
        # Parameter provided, delete from specified table
        delete_table "$1"
    fi
}

# Run main function with all arguments
main "$@" 