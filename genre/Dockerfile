# Use the official Python image
FROM python:3.10.14

# Install system dependencies
RUN apt-get update && apt-get install -y \
    libhdf5-dev \
    gcc \
    libsndfile1 \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /app

# Copy the requirements file into the container
COPY requirements.txt .

RUN pip install --upgrade pip
# Install any necessary dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application code
COPY . .

# Set the entrypoint to run your Python application
ENTRYPOINT ["python3.10", "genre.py"]