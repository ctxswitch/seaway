from app.ingest import bp
from flask import request

@bp.route('', methods=['POST'])
def post_genie():
    print('received request')
    if request.method == 'POST':
        data = request.json
        print(data)
    return 'ok'
