import os.path

def getdevlibdir() -> str:
    return os.path.dirname(os.path.realpath(__file__))

def getrootdir() -> str:
    return os.path.dirname(getdevlibdir())

def getextensiondir() -> str:
    return os.path.join(getrootdir(), "vscode")
