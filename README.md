# Introduction

This is a command line tool that generates the same unicorn avatar images as the [Unicornify web service](https://unicornify.pictures), just in better quality. For instance, the unicorn avatar for the email address mail@example.com, when generated by the website, looks like this:

![](https://unicornify.pictures/avatar/7daf6c79d4802916d83f6266e24850af?s=128)

The website returns JPEGs with a maximum size of 128x128 pixels. This tool gives you any size you want, and it gives you a PNG image:

![](https://i.imgur.com/NvySwQb.png)

It also has a few nice extra options.

# Download / Installation

There's not much to install. You can either download one of the precompiled binaries from the [releases page](https://github.com/balpha/go-unicornify/releases) (the zip file contains a single executable file, called `unicornify` on Linux and Mac and `unicornify.exe` on Windows), or you can compile the program yourself (it is written in Go, a.k.a. golang).

Please refer to the [LICENSE.txt](https://github.com/balpha/go-unicornify/src/tip/LICENSE.txt) file for redistribution information.

# Usage

## Specifying the unicorn

Unicorns are always generated based on a single (hexadecimal) number. Usually, this number is generated from an email address (by taking the MD5 hash, [just like Gravatar](https://en.gravatar.com/site/implement/hash/)). For example, the number corresponding to the address mail@example.com is 7daf6c79d4802916d83f6266e24850af.

You have three options how to tell the program which Unicorn to render: With the `-m`, `-h`, and `-r` options.

    ./unicornify -m mail@example.com
    ./unicornify -h 7daf6c79d4802916d83f6266e24850af
    ./unicornify -r

With `-m` you pass the email address you want to generate the avatar for, and the program will generate the hash number.

With `-h`, you pass the number directly. Although both `-m` and `-r` always generate 32-digit numbers, this is not a requirement. You can pass hex numbers of any length to `-h`.

Finally with `-r`, you tell program to generate a random number, and thus a random unicorn.

You must specify one (exactly one) of the above three. All the other following settings are optional.

## Output file

By default, Go-Unicornify will save the unicorn image to a file called `{hexnumber}.png`, where {hexnumber} is the number the unicorn is generated from. To specify a different file name, use the `-o` switch:

    ./unicornify -r -o awesome-unicorn.png

## Image size

Go-Unicornify always generates square images. You can specify the width (and thus height) you want with the `-s` switch:

    ./unicornify -m mail@example.com -s 1024

will generate a 1024x1024 pixel image.

## Shading & grass

By default, Go-Unicornify generates unicorns with some amount of shading, and with grass on the ground. If you want the "classic" (legacy) style images that look fairly flat, you can use the `-noshading` and `-nograss` parameters. Compare:

![](https://i.imgur.com/KpQvoZ4.png) ![](https://i.imgur.com/sa9pfwu.png)

Those two images were generated with these commands:

    ./unicornify -h b50eb7b293596008ecbb108815f82d31 -s 400
    ./unicornify -h b50eb7b293596008ecbb108815f82d31 -s 400 -noshading -nograss

## Transparent background

Unicorns want to be free! To render generate the unicorn without the background, e.g. for inserting it into another images, use the `-f` switch.

    ./unicornify -m mail@example.com -f

## Zoom out

On some avatars, the unicorn is fully visible, on others, only the head may be shown. If your unicorn is not fully visible, but you need it in full (to print it on a T-Shirt maybe?), you can use the `-z` switch.

    ./unicornify -m mail@example.com -z

You'll usually want to combine this with `-f`.

The following images show the effects of the `-f` and `-z` switches:

![](https://i.imgur.com/n7RgpHB.png) ![](https://i.imgur.com/vMfnCCl.png) ![](https://i.imgur.com/Ns0QX6Y.png)

The three pictures were generated, in order, with the following commands:

    ./unicornify -o example.png -h 772b4415cd38c3af6b522087a6359194 -s 200
    ./unicornify -o example.png -h 772b4415cd38c3af6b522087a6359194 -s 200 -f
    ./unicornify -o example.png -h 772b4415cd38c3af6b522087a6359194 -s 200 -f -z

## Disable anti-aliasing

The drawing algorithm produces very [aliased](http://en.wikipedia.org/wiki/Aliasing) images. In order to create smoother lines, the program behind the scenes creates an image that's twice as wide and twice as high as the requested size, and then uses bilinear interpolation to downscale the image. If, for whatever reason, you want to disable this anti-aliasing, you can use the `-noaa` switch.

This picture shows a detail of the above image, magnified four-fold:

![](https://i.imgur.com/z2cos5b.png)

The left image was generated normally, the right image with disabled anti-aliasing.

## Disable parallelization

The drawing operation is parallelized by default to make use of multiple processor cores. You can disable this with the `-serial` switch.
