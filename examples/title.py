from pathlib import Path

p = Path('.')
dirs = [x for x in p.iterdir() if x.is_dir()]
dirs.sort()

for dir in dirs:
    folder_name = " ".join([name.title() for name in str(dir).split("-")])
    print(f"## [{folder_name}](/examples/{str(dir)})\n\n")
