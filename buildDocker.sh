TAG=`git describe --tags --abbrev=0`
if [ "$1" = "push" ]; then
  echo "Building version:" $TAG
  read -p "Are you sure you want to push these images? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]
  then
    exit 1
  fi
  docker build --pull --push --file nop.Dockerfile --tag micahparks/magiclinksdevmulti .
  docker build --pull --push --file nop.Dockerfile --tag micahparks/magiclinksdevnop .
  docker build --pull --push --file ses.Dockerfile --tag micahparks/magiclinksdevses .
  docker build --pull --push --file sendgrid.Dockerfile --tag micahparks/magiclinksdevsendgrid .
  docker build --pull --push --file nop.Dockerfile --tag "micahparks/magiclinksdevmulti:$TAG" .
  docker build --pull --push --file nop.Dockerfile --tag "micahparks/magiclinksdevnop:$TAG" .
  docker build --pull --push --file ses.Dockerfile --tag "micahparks/magiclinksdevses:$TAG" .
  docker build --pull --push --file sendgrid.Dockerfile --tag "micahparks/magiclinksdevsendgrid:$TAG" .
else
  docker build --pull --file nop.Dockerfile --tag micahparks/magiclinksdevmulti .
  docker build --pull --file nop.Dockerfile --tag micahparks/magiclinksdevnop .
  docker build --pull --file ses.Dockerfile --tag micahparks/magiclinksdevses .
  docker build --pull --file sendgrid.Dockerfile --tag micahparks/magiclinksdevsendgrid .
fi
