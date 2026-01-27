# Pull base image
FROM nginx:latest

# Set environment varibles
ENV WORK_DIR=/etc/nginx/

RUN apt-get update && apt-get upgrade -y && apt-get install -y vim nano

# Set work directory
WORKDIR $WORK_DIR

# Copy confing files
COPY . $WORK_DIR