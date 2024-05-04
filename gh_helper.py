#! /usr/bin/python3

import argparse
import json

BUCKET_PATH = "bucket"

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Github Helper")
    parser.add_argument("example_url", help="example url for downloading github release")
    args = parser.parse_args()

    example_url = args.example_url
    example_url_split = example_url.split("/")

    version = example_url_split[7][1:]
    url = example_url.replace(version, "<version>")
    asset_name = example_url_split[-1].replace(version, "<version>")
    owner = example_url_split[3]
    name = example_url_split[4]
    res = {
        "example_url": example_url,
        "version": version,
        "architecture": {
            "x64": {
                "url": url,
                "asset_name": asset_name,
                "extract": False,
                "bin": {name: name},
                "folder": {},
            }
        },
        "checkver": {"url": f"https://github.com/{owner}/{name}/releases", "pattern": f'href=\\"/{owner}/{name}/tree/v(.*?)\\"'},
    }

    config_str = json.dumps(res, indent=4)
    with open(f"{BUCKET_PATH}/{name}.json", "w") as file:
        file.write(config_str)
