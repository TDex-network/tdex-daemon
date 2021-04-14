package swap

import (
	"encoding/hex"
	"testing"
)

const USDT = "2dcf5a8834645654911964ec3602426fd3b9b4017554d3f9c19403e7fc1411d3"
const LBTC = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
const altcoin = "913df69e031ee10891afb44c0c3019ca6c43ab7b453534b642650f594ff38578"
const initialPsbtOfAlice = "cHNldP8BALgCAAAAAAHtRVE1BkOnL3GYKskZnbmT/4wjiX+golY/TSLwpcWruQEAAAAA/////wIBJbJRBw4pyhkEPPM8zXMk4t2rA+zErgted8T8Dlz2yVoBAAAAAABMS0AAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkB0xEU/OcDlMH501R1AbS5029CAjbsZBmRVFZkNIhazy0BAAANnXmIRAAAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkAAAAAAAEBQgHTERT85wOUwfnTVHUBtLnTb0ICNuxkGZFUVmQ0iFrPLQEAAA2kdavwAAAWABTFKMruAFIAZXwEtcX2ows9PlHleQAAAA=="
const initialPsetOfAliceLegacyInputs = "cHNldP8BAKICAAAAAAFoSwHVJprhmlC85D7td20nvB7E+E6rl3Coz1MH0vhUfQEAAAAA/////wIBeIXzT1kPZUK2NDVFe6tDbMoZMAxMtK+RCOEeA572PZEBAAAAAlQL5AAAFgAUDTRn2XMDzmSXRfEjHKoR1ZfZK1MBJbJRBw4pyhkEPPM8zXMk4t2rA+zErgted8T8Dlz2yVoBAAAAAAAAAAAAAAAAAAAAAQD97iICAAAAAQHtHNQv0Ed/dBP8z35E0VUIOmimfuHETQnIvoaRmCYGJQEAAAAA/f///wMLvVzlYxHLBkFodKCMZS9nv5OLZ4ZjUUwNGwYMWLnvaLMJDRuoIiCc9emlCt7R2weuxLzhUg7X4pRl/x0fnltG1z4Deo2CPmqN8IDXXc/pvzY6SEzIaNX0R+yxvvEbN1euZHcWABQ+EVssOFPxxm9zH933CgDKnt+aPQvJpBnGt537wLBMDXR58TnVf27uroyl02T/wNiCjH784ghts+IqM4I6rRuj1TlygABN/gb8TkocHUrWIWvL17DhMQOtRSP9QD1r+hD8Rh+q5eq/4XKrwb7UntUzbzrJpFeIJBYAFA00Z9lzA85kl0XxIxyqEdWX2StTASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAAABN2AACCAAAAAAACRzBEAiB8lcbE1pC8Uzr/m1XkagkpXGTJxwy0liB7v5dt62iPSgIgVOnwQbosTlDJaA3xrdjGr4K/t1vom1f89Q96S7YoAC4BIQIbW5lFZtF55m+uliq6USBVkQXwx4a0QA9Ru4C++1H4rgBDAQABTjIEcUbJXaWaRWRu4uV4NVLUZlon/9Hq5QAc0Ig9BfFSPWp9BdgyGQS10PKjfqLknbr9XNqf2Qj1RW+Rs7Ct9f1OEGAzAAAAAAAAAAGXhHEBhxEb9AeVjR/c+nyNSxIGC2iz65P4omB6+lAB4YWaMAtvS7R0yxMeMTH1z0virWlu2dLw7cr9044sJeYtJa39C+Qsceh4gBMoVOpH2ONo/bycou+OeV9bn1HO1X7HVVwYpCzdHQQqwZ8MrNSpWHwfKiqElZaLRukrhvgqraTaJXjFeUIFnISJwjthJG+IfINIUSr6B0qsV84LFbcEHc3uhnrfla9PsXX8utaNVBAHhywYwX2Pk0/jqRiqbBYT0UfDYzYtNrOiKdVaKjQKcTkqLBvjumOgaRcyFx7Sjo5U+i27/yXRjJt/fYHWGfvNKwpcTd0Xejk1U6as3JjimDoxBO2vtU+vHrBxy6AQVpgkNSrYFrtBp/anjzPeLLCUcpY2C0BzFgQIHFECGyqHfowl3OQnHnJShRSaXj5KYT9o7q1mf60YpU3Wxd46oJe8MQ5qPpWHpuan2Q80E4nXkHYvwYW0RAB4sWryUni1y1Zuv2bj32g4Y7+KGDYvLJ99rOygC+VyQoE1OsWV0tNXYC2A5L3baa2c01rxK+NLKrhrV7sMJucfdpukYLfrWSIVVN0aDHcCPkF0Wyi9hvyuI60Rzfwi2Oml6p3dj0pnMJCdi+d2QOfNf59kD+jlp2HnRyvK1mxt54Ng7SkUh8R+llWiBsLee42jgn7w+GQ7ZsRc+Da04coK6pcm71wkwUA8yIs2OezagJgsyruZPKx7io+9uu9QT9FkW5Hbu2XMugdUw8LbHh8hQq5icsoraP5DIMljsdvnMyTmdqGU1K87hi3z1StqXQEOfd+XvaWf0f4gdC+My8DszVDdCWkeeU2K3u2N8YNfZXvEgeJgYaTAU/pztizsXkPiYGhM6QPFMZ1M1ShRyTF2neizt/5WcElSMLWF5pHTGLSZRh6nFZ94bAwkVNPhoD7EwXZQvcJEgPFzbfZv+yi7jN+n3NnwyctdU6BqbYPKpUj3Ptld4VZ/YjRH9GcivUFDvjtnJrPHGxHIrm3xYwm285y/SyljLUnBTr1mr36SsJUEDtyrCO0CaHAuMKxXfhweiM93Gzm3wOTieTcbtKA0CVm3i6PpMfsQdkHtZREYAi9Rvv65zEKqg6ONwjodaZV0vO0urQVIzHgwdPCiPCceYw+oFpyME6W+dHkIGuciW3NKGr6GYrXL5K05myJbfp834pr5fGJJEA6qweEgfNOVdKs4f4JgAvzaCUuOytIJDMMZR0NqC5DHtgDYcgtj1CQEzwOD5JDookYcK3rFAfmp6xB30uWAhgvIwzc9JDWSrmd/LqPp23bPoJSA/tKiIhK+Y42bKAPIFMqgWQ3MQfG7z2BJwtyWsfZ/rsBYB6n1o94rIOlBVs5y4JxKT8jlYELHAmWLCKYHvXgjf4ZTRenAJLRCNi6dB8LjeD+KGfxD1ObNgoaQbkq1rAMQrnG31xCL10dsrjChihmDpr2tb60ziRoLwQTGi0534WIb9H99iBr2tUMs/KNxrklx0qgKn7nWJfCRcefjEGCrIJbhDBwJQ0dZ41soHL3bCjuj5aOMBIUiURUBCJoy13aL6G48El4oJ6Tlse3ZpJShaR9vU3qherlvWXEvmc3mRxzIjKi4ywEKQTNxkzsMaZB5eaY59WEKmcWKEZaTp0oAIOSqyRrB4ErZs0ybHduzUzitR28N7I2erTfeB5BN9px4TMk9HirzEYNMgqGbcAKwj89X4bZUaDlgN8GURboCAx03hWvoRqnbKmvudvbafNNd0ndo8mfszSdywWMgmxa4HRpjr33TCUqsNEPSL29D9mGGBbT3Sn82uT5MqcRd3xXX5KsqRuSsB/wCYJ2IGWSlKJA75BSlGsVP9rMMr53YZ7eFwWY2cz1DM9c69i/xr+sHVs7SqlUY9PmDgKbVR/MbPPpZ7gDg0yMLzOTFFnJP7+thSASeFKArxlej6/9xHcY2I1qqex9+B9jZF57firxX8BdYCwPE6LEhiJ8PAq4KqU8fkM4WB4gDu/Z2kDAnF8IoWnUSy85TSXAyQzSWhvjfKQ7DfVAc1+Xq0nt8ZijYixS421CZlarRX7C9gN0GcKXmS0liy+ljDDP0WEpkX9FvcGEpi8azExodw+pxxkPcab24z+B1+j9F/Al/JY5X54qaJbBwc/DcA81FqXc9v9evSQt30qe04BNTtd5ZZbZbe8Z/K5gIYGxXdzAWBZvM3GcvOc8k4npgCdAfOdooVoRhcVmz2Ags0y54YwSwhzh62bGNBADyCHil7sD5Qr7FLI0UL+UnWnbs1+UIhViDEMsHjixccIw/ci7LoRji+usAb4ryJzVKnJCuLZwR0GkOpKuYGigJ4C7Gc/r7m8z0YUV0RfkLSN4GvFNFWOmOQNuMFZq1UkDGZNn1PiuB2L7C5Ee3nEJk9yG8N/muITYi9wblYQWqkaT6k1e0SG9+js7oG4lsWl3nPVMkbGA8g1z7T+UjULm8hz0dyxam6ONXBnJtI5bs4+cenQl2ksc1ENliHm5FugqQri7OX+EbZJGI4R//WEutH9n7/4BxuB23Px+yU0pFVkRdhmh0A1bBfjaMMRtmJy8DHAgQtAlf+KMjiRBv3avPv3bJkrZw1dDk4mlh0Mur88qSx3jWTc82Lha1cLjyjf0TXCaCj6CDpyx/20bvafLoAsyt26u1aDPcTxZoJhDWDuR1Dbaq7F8miCLyVY7DvNgRnHGf2raqCrRQx80ObDtUC3yBMpW3yjypnvcVdB+lGrbtSczUOwuTf3xWjPiSkOQLm0dHoMuIeRtFDy3A7uaI/ewTDu54rXVifeb5u7IidLEfVJIx/4hRKZT3vOqpqIixOTFZgOJS3Eh6v61rDbaPprCcOjm8m9o1Mcdt0esPhwTyvg8pjLl9pBVZsPeilkwnXXG3XEMlNO0JgGHcp4KPWylRE1y7wJuPvtuehSJbsGKs25tyL4uZENmSYUF/Ei6kbSLaqt6wqCreBtuwZlpOgMPox7jSj8PllXSCJQQyFSL1Y3bzNQf/72svwoM3SmuU2q1zDeIF7cwjvo0PNNHe6X9G2sI5eF042gLcRhJb8/w+IQV28O5eT8JjTd0ylckKmltMiiQmsYIr4ZpJq8AS4h9ijAOmLgfLAH2mFW7xWbsed2dnpzRmH9d/0I2L6/KDXvKXaZyQb+k3YUlRGWzW3uvC9fV3ws9p7G/m7FM45G07GswJJuH0wGNCV9rXlSuAC8OGXSx61g5S6/HBPzhqLkq8Sr0HlDJ/Gh9EJBEMMMDSsIwwod8abb12yorVhQE9yyIlth5SpF/fD5t40gBWrDX43PkVpgFmmTFjmiSUcI/aCTPvlgz8E7LexB3o2/84cguybfcdKFnXRLzpiTQDZ6O1FjKPMn2OrnJXoX9o2leKtCxN5esSJ2waXf0vQCpa9Vo6EEzVQo3qP9X3spuCkZOS3x65wgkGH6r1ztSOpkvUZeg3gX1xDHf9heB56afIkmK1CXoxg+zIHpmM6IMHdmdo+1OdWZPuReAjuqd5Rj+BLzTAsfUXCkQSHQ2aVR80jlsYWv24Js1Q2zzCE0Bf0K+ZWqM1Y4k5+OCScdygBAILY0aczVtcJRXiN5+6s4o8cXeCxvtHkLvaLhFw2cLhsbwDy/JjNzaaT89lPEleQIB5eP2hFWJZvaOd+uOrv3xEiCKtOq4+R9HhFYzPmmTLojwPArGC/eQkEM7OrynkjM7S+Y1HP6d0qctuiD3DByBv9v11mZGH8nfkeyk8PR6TzEOmLGpmhu/VTaqsGHwVK4huzWiBQzv3OPuAUklsB5j1ol0SVyj0IuLo9bUVxr0+8UA1D4xiccb1lsVTOxfnCuuveCjSyzOnWlfRbeWJxu+nV7jA6oD/bUGmy20wULzeh/lek6HGBVrM0+nUYfoy83GSJrXZc5G97iCAY60BjVWqgYhb97Cwy+09UYHUamxZAwvjCA2mKdUWYsNE8Ecbp0oznbQT6nQnzpWSqtop+wEsV52CH2z9VjoCpriQ/3zyExiB6ErQNxtNKXyLgBqyjaQXSApSGJAUSel/7QGwE+XnJaRC6Jt1koVmQ9gRFfZmrCRWXUaRWle3acmIX36oJVbmhIx7io+dtDBIQkJ9kfbg+KyhS3uSo2npUydxqIvZvHCBzGhNle5zR11JIOe0yyeUWJVyrzMK+3WgncBbXxQN0+7ozrSVKMXqAPmYm3p5eIMB2gsQJ6E+em+dQMOSUVzgsr7ShTrLNu9J7H938lXgpN1f4g83GHrvVCa5iTD9cK6PBrBhT/WkY0i1nBsDsSPsxsUsBkIucSr7rDgFCIO62gs2Nh4cvTUg2xxo3nIqT3R++hXV+eaWpFsRati4F266XVDk3a60QRhhjYnyYavsEUD48GvMOW3t+qjgJrmxHM2863gIyDEkEXrEtYsQL0T2Ryi/4HNt9371RfnvW/S9Xowdp0dOvsR8ulYaVHkVhutJNWSEa1QXNNork073YNFvIR9qwAVMukRx7PfA/kRzIHf9Q43I0SyTq0xnOQzWp/+MqFmVIpieEOrADrNIJFcI8xaiv1ZquFmawsye+4r8/3agpIxkGaqzsvCPa1DRy3fK+hpndnfFx44plqUl3UMdA6/Fubr0CsJjIydI3WBPoWwNPAkiYYVhV4yKUK93N//vxqUDWwCGwaoVH4XUngFcMlXzf10i7OO0PXdwFep43zPEO1mybofPa3mrsIq8xxfZoweOaGjQ5oao4OvLOAV5Fjfsnx8GFw10nOEY5CdsW08hEwb3tV4E8Vy+59ZAH9v8jcJc+KEW4vnMFsdOnkEbQRgNHfm+VpLpb61tkKNBJQpErfHgHXcyUfiGsEFzDKcWZWhOEXaCr+U14MI1T8R2O9OinRkYzeSG1K6og+Dpc9TFqzlnd6r86GvMw5I/bDGbrBNSRpf8/ENBmSy1E2kgU/QsZZS/GdAOt3kfdG/EjLj9e4xE0ccIZvh9Q25Z6GfRcCt0Qgz2b5W+VOmCfWMXWmE5kJq3JwsuenpaLPS5VsUEr/BKWIJ/O228/BqoqRlTc1sjk/Y2vfvXOJU1Tp6ImBBJbHks6PpHdaLTAGHhlRvUm3wsA5u+m6KclzHiAJz5b6m23Z3929DbqvJstJzefvV22HcfBhLctiW3O3DLXOfcQ/M2EJfH95Lo+p+Czs/zdEsW17M1FqV6bP6mjU86nKhFIXoa+uo4vdDZynixaiIN3DdgNJDpHRipywmAXubN89u0Lu4D5RpXCiuJMfdfKpLc/Ssu4YoM4/d++aphrd7oA9Av7dG9ddP0ersHk2hBKXyILfhLj7f75n0NXpic6v3/93HVDx1dfs53qprPXZwhpt3nxoMh246JxaiX9kZaThiACbJUip8Q4KZFGgxG0elBkWmdiFLsKNhATVs4lFzIXgEI8lkxD/aNUNtMqjYc/BC+ZgCe9+um//+Jl9V0Z1S2Q2D7PuuqpFRPnHB/KWoB1V5Dr27xEorhOKNiquSTKLnMLX2GzTt+DV9wbQbgBu8o6wVFY1H3QBp+U/HOffz6VDV3Xp8S0k5Dfw9sbBLcXyGIIrWe2rAgy9hDAQABFIdyH3Ci7de0C2C30SnB6ZHsu6ogJFYkd5WIN2mKmKKLBwmTsuwoKZx7zyEtYJ/DDJa95I93nkFP4GFt+vyhnP1OEGAzAAAAAAAAAAGYdAcBYXoxQAXU/ulxMKr49SiEIlOg3z1uFXXFOg1EF1DvrQmBG4OY4RSXyNG7q2BVkfu/Yw43lXIagxHUfCgZ0ICkadFiA0c+6nzvztbNDUHCjZVy0wZBaSede13khm/lJute7yczH/7qyqobjTqvwaqp7v2Og2o2ZueWIKAnUKl6PtoVQAt1X08TZu+3fwNP1dV5IQtuPsOwnEmLn0v+Ex74YiX66feBWBXesBhm+BwiMazjDOtqJlaS/jRqcr7va5Z7GCaKwP6sQO3LS23V5NB0PODqSUuOb2vaMBpbRGIkdm/ErWkaERYUjDbBf8bCkv4203asQ6FQMt6mbUoLAWu/f/4kOTaWEInY+ySa0Newmq/D3ISieGe+AHP4RiYQFaWcmURJ7JAaWq7yV1FmzPIRDOagMXAC02vaNvNrzWQHvZ5TxPBCeHQiwCsoLgka6G0UcdBwYPIpveRLzeAUD7hxq3HI81YNrFC5i8u7K6wc9osPwu8dbyR8nn+NDCk8ZXDRTp8naA9LcQPNBrRj25Nh3ZvYA7Zz4e7wm5Qey1zqkrS7mCkLm7g0ogoiO90NIzwyVkH9MOKTuZDIRZRatCmlG0AEKXxIv7an7C/shfjd3mN//JBkU1FKRAg88YnucNHYcNFaI5JMcvDrR3qYvZ0m30J2svCy4JNupHTU6cmYEcfmvhF7YJyOZjLLJVJu2Gd+eOkOehYrStQGMPSKw6fPeAQYodr3YpEMZETALEGzPLvXR3VZT0W8VqrMHFwbSz3+FqkyKH8svpid5/YqMiIjbg3YBFWMTZCdo+Jg3BwDh7f7sr3Us6SBPCTM5IRkm+T+Cc3B0w9mDJsji+Mr/+86bUoJm3hXHYPgaySF5JShDxy8/GhcPRULxZOL8vMQz4hxzwLcdjH9CouCiWfIru0dhRwDA5lgLxBRe2LhywKPN/0ErDDpU5E/bPwKBT1BkNY8fR/Y0Z2b9DfExQZdPvP/OhrhsX/VJKWdch8bIjhOj7m15L1MrFSmh4ie+SI7GUK05CXXhv6IN6kxnG5/Eg3l23hXQuqM14bkIwfVsD7NJSIhW1JLFraBOS3oORoS3HIlP0kyhxIiXhR7DWqoKupmc8TphFEjU6BXFpLD/aEsmFb2oMA2H05y2nhr4VmB3VNu0X2j5FBk1o4f6E1SrdqlpiTk4IkbqXGhd6z9DMif/AKQYquUVqZMBzMChpb4r8AT10zhJxGYc6y76zZFdMtu6Z1tbi0kTHzYrjCmwkOvMgVcpPg6NSXawQP5D2T0dV3HgpE2uXgs0cKyTqBqRyH9aogkhLabTECAPEgASfr1a2yB17Z2U8DakBnOAwGD2P5CeHYMo1ymMxIj2GycAfIABqKL6nz27hkuYcvac1VtRTDP5WktpbaBJkS6kd9wPAsZcZxGitipUVo0gGHhVDdwrYT03OgAx1rxr5c3lhqydSnIiG0QjI4UyLHzFNBy+JtdEDK2VR8yUl/NrA/qnz9iuXqxgl9r/G3KWxsanwaIeWOKjQN+KOsZAPHE/EhWt/grnxZ6+gVqHtW2Zn/NfNT1otLnLaQZoA92rUAm7oDZ74cjSJiX3i0YkbAuZ5o8p1WSHZhDy09OfGiaoBSWno33wiZqTbZqGhErowJQz1pac5fK0zGjzE0LXQgw5jw8qRSWfQT0SqSIlKYRPuWnYJEbelXeH4gfjcXdB/f9qc3u6C2Ryz3mzyMTC/qJ/BBhps1GgCpHDTlC+zdIKahEmLzPFgjviJZi5VFXQrTRoR+LBkPtqGdIcF6YTHINVDqcUPP2Mq4oKaBSRUODEMnkioPXX1UWt+hIhtVvNKj7YnI78WjMVthdbdUSxRXvmq/3evluyQEDM4BCiAH7S/nP3teYU6RYeAeAHL0Kb5FxtgRxrDiXGDed4AqG8K8yf+bl8QsC8JVPtTfZzFtEa453vgGnuS+PV3qo7qnYoG/s4EYl1indNv9GB5mbaWtZJIAIWGV8rztBGcP85OpVPjUL9/bQIvnM/dQJM+rSKkuEtwq795q0snNMApf3n1OddsBkZ7m8lYwcgf5pO6Lc0R9oySDoZN9imwi3jCNuWSTLjtokZVo+oofRfHSBTuFaL9lEGpcDOCkLqwEf+hZVVxjuDCTzfxffYzAe4BNVesMuIJg8xhqx0oWGUhcFJFJjq9gh12r4GFfqLSWoxcbuGAzLHVSYMLdb+naKBMo9GeSyzyKJyQ/ftrPRj4ZIhX4Ew4JYuzp+55MHd8mzEVuIEwPl8SUn5EA5Ej/ecTlycEW06r3ssMudBe79YwVC2DkuoZXRvAApBkfWQ++BMi2nOId8RlRPf1hhB3yY6sBzoE1x7DhgK9kAObbLwfs5g0zWvEQBUiZChCOBf84HLZIPLTZA/vQJ6JELqppxrSYuSgNKC0xD67lFzEvtL1oDZWuoW9blYEhqUsxcJA4kQYVShct5w+3YtxeIQX7VvMaDWU3P1iI/AKcw64MlmnGYLWZLlx1q412+dkK7hw9Hh8BwjFzvSQv7YmgbDjaQjnNQs8pIG2RApTsSHfPkvtZ2z9fmfvHVeqnlwuJyz7gGNpacEyhU41MNkc+TpqI4VNKuEErWfNeejUKmDZrjDdZ5kmyiRUM4qO1Fk+N9vXIKShogkNs1hJQje3OW9rczVlDKP6Et8rijc0+mb+RnGLD+0jNw2zXx6WrP7YUWEoaun/nJAEgBzGVaHXh3rvx83902wQm30XLa4JlCOdwbggMzgJh+PHGS0yU96o9jKPu91Ol0u9XTUbUZz+0ikTZng19z5DmJ9tMVuWS4uo+9V7biiI/P74CANzdQOeLgwUi/8KFKPFwZekOi+MsMsEKzcpvfRZuZn1qxcp6zAabAuwqXCg1C0wHMjw0xfJ4MHqC83K9cDO3RaYwWqXXJ7+4ReWAxh7Nbjb/YuG4CeNzSpohVBZl4MtyO6dup6KsFSoKDYdTBLy82KEpLVaU24SQoPfA2FW8+a+0y+D1Rw5PYWFdDcRW9mr4K4TjETvRbXDQc16Q4WEdhkQl/J5mmzHgbWLRtg77bwbJCM5JrJVwiCh6bnSpPyMf4PRqXtGyWHmfDWC7Nl5uZSwGOXNuKukRy/FdoxrKFCQPWBeoZKdkx3ygzbOY3x4O/QUUyYpCmjloouYiMESH1qRvlrYGrnUIRU265+9TR02uGGnpONJMVaWgkFqMpTpfJrRsaZnPGgKlHFcagCAs+DOM7wxhDmhhphJsxSD3QnB7Ibl6JBKhGxgWrM3UZm4bkO8uI9ZHI+PJ0Wy3LABxqZEK564kX0XKsqblwKHBcoAKUAcZ01pAUVkJqzVBi9RG69mWbWLtQpzbA9/Ak9uyOWFDEbC4/VqB6xjM/FuWCsHOR/PAV1DfrdKyj17oAuQB1ZVCZMjij7fTb/vt//GGThTHXzr5rsHDalPmAST3EYdNulQVGWFFrtUMgg6jGRNkqkuL39BTPFHMqLEhJGqxwP4mE65rii4KFm85MfNPyWXM7ok+L+R/ZB9b/YpctVQ65TpObBpXV2l67ty5P3vIiRomO8PEE5oGa39MwizN/uEYsbV+8M2MuN3VEgxgFMTz9zslSb0i1O2ZT3hK28ODNH6g+4FIdVywwbOwgiCYkAsj3NTz3IJKa8xXbdJRToNTPvzSJsX5ukPVDaEA7LPZZucIEC9wdJCJTzTlDJFpnKqzuNfmqRmapZ6irTJj5g9/Lq3WiTa3blmbR/acxXIps6Mu0O2q/eDam1S9wMprLpWD0V1BNDlOn1WSNJWxlEnTaRHQJpyG6+Cfg+vvyKew/eOKd5VOJ6J3MADzY9bq+197Ly7klWiqRU361QINgKkLUjIxDrKuZDRI1AyKJd6+AL02vnwGxBr7OAiZE+VN23X/U96D0UB5exI3h7wTSgG6DTX/0YzV48ZOVl1yIUaM/CYLvHFeuc6shdHOu/n0pj5TFTsV3C9aygVQFIL1m+ia9mp6163LAYZ3v/sgRd6ZIe/MX1AHB0BQLv/tu2KesPNK7hHTZzS05a6+5IMyXEm5oElCOoOXeF8eCVMkHtPLzmbmWWP1KJLlGI1iy1F+43Fca3Fi8NdF4CQRd9PSSxKjUEPluYRbDqyK1ECSBqQZUaK3vUdjaKJK1/OkoYr5YWj34vs58Rug+SouCNeVt2x1ATKqfgdfbBxPArkTlzW31v3nfJomEoGeay5OvT+qFa0eFXp06nXsCr8w8qIj7q8M1XT4IcWWbjC0ZuXR/KsSjqK6wI/EPvpMOjZWLVZ6Rg3h+ndh2GLAGVyYpxkmRxaEL8pPOU+JO1+7jkLaWrl5QOn5oO11iCXbi/Ul+KZ5ju6bhXRrOj8OIM7Tw+mRVMdFGvemE2+RPtAtMumdwtxOC0ewee9BJuQZi39cmukDn8+Z1gEwyNXDxEEBiFdlZSY8IxSfxE2qY7Xu5d7jhqx9CONhogLAvzypNzFgQsKC0PGPxGGJoMLPnNqqHVr4TNl3bm8lQZ93QcNQODRePd81+ATLsKIoK9RDyMG8DWV4/H5aU1jEnqdWbu8hjKtRpP2BTEuOYxzMM47oREmtiOXx6KY5OOp2E9PSIv+zc2ZfaaINipP/O8qvTClA8DLloOoaHfR443Mr36yg4RN6dDD9VqQB1Q81Q/wiKNsMt1SQ0xJrHt65y+j9ZdgteiifZfxDBXkMXGQRYXXq9LrCV8FSuA+RM9JiiXusUakbx0Vydhg8AmGkuMoq8t7eeUdHVmtqxu273Gk67x+3ZKxy/Jwsppd7FyVvDCC8D1r7+kgn7KtJk2GTDmZIY1Mx2YkhJ7Zlt1OCdzZbLPfIqsLWXIE5bnKfKZsJFLywzNYBVJpLCyfag+00o9q5vveUoB2gcsKsO61BF8Nilat3G6IfRSuQCfOQSYm4/iIa5uMgJqK+wKkj31vm+wmqdzYlcRfgszFI7kldfda5PiI8b51V+341ZFsDWJcKKdgf/JlaXay8/kJ/icPGxcBYpzmckdJwHGM3kfXu9T4oPFKjU1UN9Ecz0ETkmuRFLbOpU7IVYa47SzjaDcVUP85FHFYYODDeH/uD1989es+be6ZdvPJLgP8UcVVocAl+KkOzl5etRNC+yLCPd5cdGwZ5GeQ9xtbSU0mYB1LAWgIKQef82Y+53x1SYJvkRc2q6g3vfRisrOmjmy7LqxWw4gdhrdQ494NyqJbgxwymLyWZInHJhwyAYb4VmeT9sSaiECUMnqKVsItAlmDDnA1rTxMoraBDjQBD9MW6/auiIW6sHE72hvgB5/h+1SySv59aGwRAJuSiv3JTDFAeknpyvkPxuPC6RGB/RwV4J4Q8051V2+ybYy7oDSJZt9N90SGpkFLNbQnglyxMffyEe0L+BNPMv/rvbT0uwBzHvGgHCiGGkDjaBcxc89lMUPSQmnDXgXt5nlMITCgnHppGVp8EZDH6aUNaKLiJTZbUG3W5IZvzH758ZeOMNbaod3qTHJ4hNTa7142sgRHt+kP7e8K2NI+EMq+bLV2L1c8LLs14/gGHdQTZCz20SFCLQUDUtibVk7dWalaoNn9njJKt39UkAAAAAAA=="

func TestCore_Request(t *testing.T) {
	inBlindKeys := make(map[string][]byte)
	firstBlindKey, _ := hex.DecodeString("9f252edcfcdff4be10fd03983adee22f6279bd235616310db40133b219ff9acd")
	secondBlindKey, _ := hex.DecodeString("35522f2cd456611685fc66f56a50c3fdf99b3b07c23a8399a9a7f3720e58b91e")
	inBlindKeys["0014c0fdf9e9bfb0d08fab79291c9b242a924e44cd37"] = firstBlindKey
	inBlindKeys["00140d3467d97303ce649745f1231caa11d597d92b53"] = secondBlindKey

	type args struct {
		opts RequestOpts
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Alice can create a Swap Request message",
			args{RequestOpts{
				AssetToSend:     USDT,
				AmountToSend:    30000000000,
				AssetToReceive:  LBTC,
				AmountToReceive: 5000000,
				PsetBase64:      initialPsbtOfAlice,
			}},
			make([]byte, 520),
			false,
		},
		{
			"Alice can create Swap request message (legacy inputs)",
			args{RequestOpts{
				AssetToSend:       LBTC,
				AmountToSend:      100000000,
				AssetToReceive:    altcoin,
				AmountToReceive:   10000000000,
				PsetBase64:        initialPsetOfAliceLegacyInputs,
				InputBlindingKeys: inBlindKeys,
			}},
			make([]byte, 12656),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Request(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Core.Request() = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}
