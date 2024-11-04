from flask import Flask
from config import Config

def create_app(config=Config):
    app = Flask(__name__)
    app.config.from_object(config)

    from app.main import bp as main_bp
    app.register_blueprint(main_bp, url_prefix='/')

    from app.ingest import bp as ingest_bp
    app.register_blueprint(ingest_bp, url_prefix='/ingest')

    return app