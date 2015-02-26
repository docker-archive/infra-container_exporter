FROM       golang:onbuild
MAINTAINER Johannes 'fish' Ziemke <github@freigeist.org> (@discordianfish)
ENTRYPOINT [ "go-wrapper", "run" ]
CMD        [ "" ]
EXPOSE     9104
