from app.main import bp

@bp.route('/')
def index():
    return '{"version": "v0.0.1"}'

@bp.route('/healthz')
def healthz():
    return '{"status": "ok"}'