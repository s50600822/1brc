#  Copyright 2023 The original authors
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

from concurrent.futures import ThreadPoolExecutor
from collections import defaultdict
import sys
import time

def process_chunk(chunk):
    station_temperatures = defaultdict(list)
    for line in chunk:
        station, temperature = line.strip().split(';')
        temperature = float(temperature)
        station_temperatures[station].append(temperature)
    return station_temperatures

def merge_results(results):
    final_results = defaultdict(list)
    for result in results:
        for station, temperatures in result.items():
            final_results[station].extend(temperatures)
    return final_results

def calculate_temperatures(file_path):
    # CHUNK_SIZE = 1000000
    CHUNK_SIZE = 100
    with open(file_path, 'r') as file:
        with ThreadPoolExecutor() as executor:
            chunks = [list(file.readlines(CHUNK_SIZE)) for _ in range(CHUNK_SIZE)]
            results = list(executor.map(process_chunk, chunks))

    final_results = merge_results(results)

    # Calculate min, mean, and max temperatures for each station
    processed_results = {}
    for station, temperatures in final_results.items():
        min_temp = round(min(temperatures), 1)
        mean_temp = round(sum(temperatures) / len(temperatures), 1)
        max_temp = round(max(temperatures), 1)
        processed_results[station] = f"{min_temp}/{mean_temp}/{max_temp}"

    sorted_results = sorted(processed_results.items())

    print("{" + ", ".join([f"{station}={temperature}" for station, temperature in sorted_results]) + "}")

if __name__ == "__main__":
    start_time = time.time()
    calculate_temperatures('measurements.txt')
    end_time = time.time()
    elapsed_time = end_time - start_time
    print(f"Elapsed time: {elapsed_time} seconds")
