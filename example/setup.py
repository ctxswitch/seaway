from setuptools import setup, find_packages


with open('README.md') as f:
    readme = f.read()

with open('LICENSE') as f:
    license = f.read()

setup(
    name='sandbox',
    version='0.0.0',
    description='Sample application to demonstrate the Seaway dev tools.',
    long_description=readme,
    author='Rob Lyon',
    author_email='rob@ctxswitch.com',
    url='https://github.com/ctxswitch/seaway-sandbox',
    license=license,
    packages=find_packages(exclude=('tests', 'docs'))
)
