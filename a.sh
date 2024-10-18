# 1. Install Dependencies
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    libssl-dev \
    libev-dev \
    zlib1g-dev \
    libc-ares-dev \
    libnghttp2-dev \
    git

# 2. Install ngtcp2
git clone https://github.com/ngtcp2/ngtcp2.git
cd ngtcp2
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
sudo make install
sudo ldconfig

# 3. Install nghttp3
git clone https://github.com/ngtcp2/nghttp3.git
cd nghttp3
mkdir build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)
sudo make install
sudo ldconfig

# 4. Install libcsp
git clone https://github.com/shiyanhui/libcsp.git
cd libcsp
make
sudo make install
