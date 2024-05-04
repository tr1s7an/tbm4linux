#! /usr/bin/python3

import argparse
import json
import os
import re
import shutil

import requests

BUCKET_PATH = "bucket"
CACHE_PATH = "/tmp"
BINARY_PATH = os.path.expanduser("~/.local/bin")
FOLDER_PATH = os.path.expanduser("~/.local")
EXTRACT_SCRIPT_PATH = "/usr/local/bin/extract.sh"

ARCHITECTURE = "x64"


def read_config(config_file):
    with open(config_file) as file:
        config = json.loads(file.read())
    version = config["version"]
    formatted_config = json.loads(json.dumps(config).replace("<version>", version))
    return config, formatted_config


def update_config(config, config_file):
    config_str = json.dumps(config, indent=4)
    with open(config_file, "w") as file:
        file.write(config_str)


def check_version(checkver_dict):
    resp = requests.get(checkver_dict["url"], headers={"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0"})
    html = resp.text
    new_version = re.search(checkver_dict["pattern"], html).group(1)
    return new_version


def install(config_file):
    _, formatted_config = read_config(config_file)
    asset_name = formatted_config["architecture"][ARCHITECTURE]["asset_name"]
    if not os.path.exists(f"{CACHE_PATH}/{id}"):
        os.mkdir(f"{CACHE_PATH}/{id}")
    os.chdir(f"{CACHE_PATH}/{id}")

    print("\tdownloading...", end="")
    file_response = requests.get(formatted_config["architecture"][ARCHITECTURE]["url"])
    with open(asset_name, "wb") as file:
        file.write(file_response.content)
    print("ok", end="")

    if formatted_config["architecture"][ARCHITECTURE]["extract"]:
        print("...extracting...", end="")
        try:
            os.system(f"{EXTRACT_SCRIPT_PATH} {asset_name}")
            print("ok")
        except:
            print("failed!")
            return
    else:
        print()
    for k, v in formatted_config["architecture"][ARCHITECTURE]["folder"].items():
        dst = os.path.join(FOLDER_PATH, v)
        print(f"\t{k} -> {dst}")
        if os.path.exists(dst):
            if input(f"\tconfirm that {dst} is going to be deleted(y/n): ") == "y":
                shutil.rmtree(dst)
            else:
                print(f"\t{dst} not updated")
                continue
        shutil.copytree(k, dst)
    for k, v in formatted_config["architecture"][ARCHITECTURE]["bin"].items():
        dst = os.path.join(BINARY_PATH, v)
        print(f"\t{k} -> {dst}")
        shutil.copy2(k, dst)
        os.chmod(dst, 0o755)
    os.chdir(cwd)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Linux Binary Manager")
    parser.add_argument("ids", nargs="+", help="binary id")
    parser.add_argument("--check", "-c", action="store_true", help="Check for updates")
    parser.add_argument("--install", "-i", action="store_true", help="Install the binary")
    parser.add_argument("--update", "-u", action="store_true", help="Check for updates and install")
    args = parser.parse_args()

    if args.ids[0] == "*":
        ids = [file.split(".")[0] for file in os.listdir(BUCKET_PATH) if file.endswith(".json")]
    else:
        ids = args.ids

    cwd = os.getcwd()
    for idx, id in enumerate(ids):
        print(f"({idx+1:03}/{len(ids):03})[{id}]: ", end="")
        config_file = f"{BUCKET_PATH}/{id}.json"
        if os.path.exists(config_file):
            if args.check or args.update:
                config, formatted_config = read_config(config_file)
                old_version = config["version"]
                new_version = check_version(config["checkver"])
                if new_version != old_version:
                    print(f"<{old_version}> -> <{new_version}>")
                    config["version"] = new_version
                    update_config(config, config_file)
                    if args.update:
                        install(config_file)
                else:
                    print(f"<{old_version}>")
            if args.install:
                install(config_file)
        else:
            print(f"config file {config_file} not found")
        print(f"({idx+1:03}/{len(ids):03})[{id}]: done")
