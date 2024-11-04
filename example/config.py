import os

class Config:
    DATABASE_USER = os.environ.get('DB_USER')
    DATABASE_PASSWORD = os.environ.get('DB_PASSWORD')
    DATABASE_URI = os.environ.get('DATABASE_URI')
