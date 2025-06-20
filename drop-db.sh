#!/bin/bash

# --- IMPORTANT CONFIGURATION ---
# Replace 'your_mysql_container_name' with the actual name or ID of your running MySQL Docker container.
# You can find it by running 'docker ps' in your terminal.
MYSQL_CONTAINER_NAME="mysql" 

# The name of the database to drop and recreate
DATABASE_NAME="qb"

# The variable name for your MySQL root password in your .env file
MYSQL_PASSWORD_ENV_VAR="MYSQL_ROOT_PASSWORD"
# --- END CONFIGURATION ---

MYSQL_ROOT_USER="root" # This is the standard MySQL root user
ENV_FILE=".env"        # Name of your environment file

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
  echo "Error: .env file not found in the current directory."
  echo "Please create a .env file with your database credentials, e.g.:"
  echo "  $MYSQL_PASSWORD_ENV_VAR='your_database_root_password'"
  exit 1
fi

# Load environment variables from .env file
set -a
source "$ENV_FILE"
set +a

# Check if the password variable from .env is set
MYSQL_ROOT_PASSWORD="${!MYSQL_PASSWORD_ENV_VAR}" # Indirect expansion
if [ -z "$MYSQL_ROOT_PASSWORD" ]; then
  echo "Error: Environment variable '$MYSQL_PASSWORD_ENV_VAR' not found or is empty in $ENV_FILE."
  echo "Please ensure your .env file contains: $MYSQL_PASSWORD_ENV_VAR='your_database_root_password'"
  exit 1
fi

echo "Attempting to drop and recreate database '$DATABASE_NAME' in container '$MYSQL_CONTAINER_NAME'..."

# Drop the database if it exists
docker exec -i "$MYSQL_CONTAINER_NAME" mysql -u"${MYSQL_ROOT_USER}" -p"${MYSQL_ROOT_PASSWORD}" -e "DROP DATABASE IF EXISTS \`${DATABASE_NAME}\`;"

# Create the database
docker exec -i "$MYSQL_CONTAINER_NAME" mysql -u"${MYSQL_ROOT_USER}" -p"${MYSQL_ROOT_PASSWORD}" -e "CREATE DATABASE \`${DATABASE_NAME}\`;"

# Check the exit status of the last command
if [ $? -eq 0 ]; then
  echo "Successfully dropped and recreated database '$DATABASE_NAME'."
else
  echo "An error occurred while dropping/recreating the database. Please review the output above for details."
  echo "Ensure the MySQL container is running, the container name is correct, and the password is valid."
fi

echo ""
echo "--- Next Steps ---"
echo "Now, restart your Go application (or run 'air' if you are using it) to let GORM re-create all tables with the updated schema."
