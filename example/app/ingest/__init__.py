from flask import Blueprint

bp = Blueprint('ingest', __name__)

from app.ingest import routes