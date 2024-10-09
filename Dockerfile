# Use the official MySQL image for version 8.0
# Note: MySQL 9.0 doesn't exist as of 2024, using 8.0 instead
FROM mysql:8.0

# Set environment variables
ENV MYSQL_ROOT_PASSWORD=econova
ENV MYSQL_DATABASE=econova

# Create a directory for the SQL file
RUN mkdir -p /docker-entrypoint-initdb.d

# Copy the SQL dump file into the container
COPY backup.sql /docker-entrypoint-initdb.d/

# Expose the default MySQL port
EXPOSE 3306

# The entrypoint script from the base image will automatically
# execute files in /docker-entrypoint-initdb.d/