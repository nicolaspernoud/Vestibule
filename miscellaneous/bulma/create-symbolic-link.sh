#!/bin/bash
rm ../../web/assets/bulma.min.css
ln -s $(pwd)/css/bulma.css $(pwd)/../../web/assets/bulma.min.css
