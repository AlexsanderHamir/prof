FROM python:3.12-slim

WORKDIR /app

# Install git for cloning the repository
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

# Clone the repository
RUN git clone https://github.com/AlexsanderHamir/prof.git /app

# Create a non-root user
RUN useradd -m -s /bin/bash profuser

# Set up Python environment and install dependencies
RUN python -m venv /app/venv && \
    . /app/venv/bin/activate && \
    pip install --no-cache-dir -r requirements.txt

# Create wrapper script
RUN echo '#!/bin/bash\n\
if [ -z "$PROF_CONFIG" ]; then\n\
    echo "Error: PROF_CONFIG environment variable not set"\n\
    echo "Please set PROF_CONFIG to point to your config file location"\n\
    exit 1\n\
fi\n\
\n\
if [ -z "$PROF_PROMPT" ]; then\n\
    echo "Warning: PROF_PROMPT environment variable not set"\n\
    echo "Custom prompts will not be available"\n\
fi\n\
\n\
/app/venv/bin/python /app/prof.py "$@"' > /usr/local/bin/prof && \
    chmod +x /usr/local/bin/prof

# Switch to non-root user
USER profuser

ENTRYPOINT ["prof"] 