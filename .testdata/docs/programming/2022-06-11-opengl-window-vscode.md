# Getting Started with OpenGL on Windows using VSCode

Settings up OpenGL can be challenging on windows. Most material out suggests to use Visual Studio. However, it is possible to get a setup going using VScode.

## MSYS2 Setup

First, we install `MSYS2`, which is similar to mingw or wsl. It gives you a linux type environment. You can download and install it from the [official homepage](https://www.msys2.org/).

## Windows Terminal

If, you use windows terminal, you might want to open `MSYS2` with it. The below config. You can add the below profile to your settings.json, to achieve this. I have decided to use ucrt64 but you can do the same for the other environments that come with `MSYS2` as well. For example you can change `-ucrt64` for `-mingw64`.

See [the docs](https://www.msys2.org/docs/terminals/) for more information.

```json
{
    "profiles": {
        "list": [
            {
                "guid": "{5a352aaa-387b-42c6-a8b3-fbf5c598ef1b}",
                "commandline": "C:\\msys64\\msys2_shell.cmd -defterm -here -no-start -ucrt64",
                "startingDirectory": "C:\\msys64\\home\\%USERNAME%",
                "icon": "C:\\msys64\\ucrt64.ico",
                "name": "MSYS2",
                "hidden": false
            }
        ]
    }
}
```

## VS Code

Additionally, you can integrate it into vscode, the below profile integrates also urct64(universal c runtime), but you can do it like for windows terminal for the other environments as well.

```json
{
    "terminal.external.windowsExec": "%USERPROFILE%\\AppData\\Local\\Microsoft\\WindowsApps\\wt.exe",
    "terminal.integrated.profiles.windows": {
        "MSYS2": {
            "args": [
                "-defterm",
                "-here",
                "-no-start",
                "-ucrt64"
            ],
            "path": "C:\\msys64\\msys2_shell.cmd"
        },
    },
}
```

## OpenGL

The basic environment is set. Now we install from the [ucrt repo](https://packages.msys2.org/package/?repo=ucrt64) some packages to get OpenGL working. For now [freeglut](http://freeglut.sourceforge.net/) will be good enough.

Open VSCode and a new terminal `MSYS2` terminal within, and run the below commands.

```bash
# update pacman
pacman -Syu
# gcc
pacman -S mingw-w64-ucrt-x86_64-gcc
# glut
pacman -S mingw-w64-ucrt-x86_64-freeglut mingw-w64-ucrt-x86_64-glew
# make
pacman -S mingw-w64-ucrt-x86_64-make
pacman -S make
```

## Intellisense

We can now start our first project. In order to get intellisense working with vscode, I have installed the [c/c++ extension from microsoft](https://marketplace.visualstudio.com/items?itemName=ms-vscode.cpptools).

Additionally, I have added the below configuration to `.vscode/c_cpp_properties.json` inside my workspace. This lets the extension know where to resolve the includes in my c code in order to provide intellisense.

```json
{
    "configurations": [
        {
            "name": "Win32",
            "includePath": [
                "${workspaceFolder}/**",
                "C:\\msys64\\ucrt64\\include"
            ],
            "defines": [
                "_DEBUG",
                "UNICODE",
                "_UNICODE"
            ],
            "cStandard": "c17",
            "cppStandard": "c++17",
            "intelliSenseMode": "windows-gcc-x64",
            "compilerPath": "C:\\msys64\\ucrt64\\bin\\gcc.exe"
        }
    ],
    "version": 4
}
```

## First Program

We, can create a simply program to test if everything is working. I scraped the below code from the web. All it does is rendering a square. Enough to verify if everything is working.

Copy the below content to main.c.

```c
#include <GL/glut.h>

void displayMe(void)
{
    glClear(GL_COLOR_BUFFER_BIT);
    glBegin(GL_POLYGON);
    glVertex3f(0.0, 0.0, 0.0);
    glVertex3f(0.5, 0.0, 0.0);
    glVertex3f(0.5, 0.5, 0.0);
    glVertex3f(0.0, 0.5, 0.0);
    glEnd();
    glFlush();
}

int main(int argc, char **argv)
{
    glutInit(&argc, argv);
    glutInitDisplayMode(GLUT_SINGLE);
    glutInitWindowSize(300, 300);
    glutInitWindowPosition(100, 100);
    glutCreateWindow("SAMPLE TEST");
    glutDisplayFunc(displayMe);
    glutMainLoop();
    return 0;
}
```

Finally, we can compile the program. Make sure you link the OpenGL dependencies.

```bash
gcc main.c -lfreeglut -lglew32 -lopengl32
```

You can see the program running by executing it via `./a.out`.

## Conclusion

We have setup up a basic environment to work with vscode and OpenGL on windows. Our first program compiled and run successfully. From here, on we can experiment with OpenGL. Have fun.
