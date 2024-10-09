# Use the official MySQL image for version 9.0
FROM mysql:9.0

# Set environment variables
ENV MYSQL_ROOT_PASSWORD=nika8224
ENV MYSQL_DATABASE=econova

# Copy the SQL dump file into the container
COPY backup.sql /docker-entrypoint-initdb.d/

# Expose the default MySQL port
EXPOSE 3306