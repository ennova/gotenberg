FROM debian:bullseye-slim

# |--------------------------------------------------------------------------
# | Common libraries
# |--------------------------------------------------------------------------
# |
# | Libraries used in the build process of this image.
# |

RUN echo "deb http://httpredir.debian.org/debian/ buster main contrib non-free" >> /etc/apt/sources.list &&\
    apt-get update &&\
    apt-get install -y curl wget gnupg ttf-mscorefonts-installer procps

# |--------------------------------------------------------------------------
# | Chromium
# |--------------------------------------------------------------------------
# |
# | Installs Chromium.
# |

RUN apt-get -y install chromium

# |--------------------------------------------------------------------------
# | Fonts
# |--------------------------------------------------------------------------
# |
# | Installs a handful of fonts.
# | Note: ttf-mscorefonts-installer are installed on top of this Dockerfile.
# |

# Credits: 
# https://github.com/arachnys/athenapdf/blob/master/cli/Dockerfile
# https://help.accusoft.com/PrizmDoc/v12.1/HTML/Installing_Asian_Fonts_on_Ubuntu_and_Debian.html
RUN apt-get install -y \
    culmus \
    fonts-beng \
    fonts-hosny-amiri \
    fonts-lklug-sinhala \
    fonts-lohit-guru \
    fonts-lohit-knda \
    fonts-samyak-gujr \
    fonts-samyak-mlym \
    fonts-samyak-taml \
    fonts-sarai \
    fonts-sil-abyssinica \
    fonts-sil-padauk \
    fonts-telu \
    fonts-thai-tlwg \
    fonts-liberation \
    ttf-wqy-zenhei \
    fonts-arphic-uming \
    fonts-ipafont-mincho \
    fonts-ipafont-gothic \
    fonts-unfonts-core

COPY build/base/fonts/* /usr/local/share/fonts/
COPY build/base/fonts.conf /etc/fonts/conf.d/100-gotenberg.conf

# |--------------------------------------------------------------------------
# | Default user
# |--------------------------------------------------------------------------
# |
# | All processes in the Docker container will run as a dedicated 
# | non-root user.
# |

ARG GOTENBERG_USER_GID=1001
ARG GOTENBERG_USER_UID=1001

RUN groupadd --gid ${GOTENBERG_USER_GID} gotenberg \
  && useradd --uid ${GOTENBERG_USER_UID} --gid gotenberg --shell /bin/bash --home /gotenberg --no-create-home gotenberg \
  && mkdir /gotenberg \
  && chown gotenberg: /gotenberg