#!/usr/bin/env python3

import io

def main():
    with io.open("/usr/include/xkbcommon/xkbcommon-keysyms.h") as f:
        lines = f.read().splitlines()

    print("package xkb\n\n// #include <xkbcommon/xkbcommon-keysyms.h>\nimport \"C\"\n\ntype KeySym uint32\n\nconst (")

    for line in lines:
        parts = line.split()
        pre = "XKB_KEY_"
        if len(parts) < 3 or not parts[1].startswith(pre):
            continue
        name = parts[1][len(pre):]
        comment = ""
        if len(parts) > 3:
            comment = " " + " ".join(parts[3:])
        print("\tKeySym{0} KeySym = C.{1}{2}".format(name, parts[1], comment))

    print(")")

if __name__ == "__main__":
    main()
