import argparse
import os
import time
import tqdm
import typing
import subprocess
import pathlib
import logging
import multiprocessing

import pandas as pd

def url_worker(root_url: pathlib.Path, queue: multiprocessing.Queue) -> None:
  """
  Clone in root path.
  """
  while not queue.empty():
    cur = queue.get()
    pr = subprocess.Popen(
      "git clone {} {}".format(cur, root_url / cur.split('/')[-1].replace(".git", "")).split(),
      stdout = subprocess.PIPE,
      stderr = subprocess.PIPE,
      universal_newlines = True,
    )
    try:
      stdout, stderr = pr.communicate(timeout = 600)
      # print(stdout, stderr)
    except Exception as e:
      print(e)
  return

def watchdog(total: int, queue: multiprocessing.Queue) -> None:
  """
  Set up a progress bar.
  """
  bar = tqdm.tqdm(total = total)
  done = 0
  while not queue.empty():
    off = total - queue.qsize()
    bar.update(off - done)
    done = off
    time.sleep(2)
  return

def mine(csv: pathlib.Path, out_path: pathlib.Path) -> None:
  """
  Scan github csv links and clone iteratively everything.
  """
  df = pd.read_csv(csv)
  urls = multiprocessing.Queue()
  for u in df['url'].tolist():
    urls.put(u)

  total = urls.qsize()
  pool = min(os.cpu_count(), urls.qsize())
  procs = [multiprocessing.Process(
      target = url_worker,
      kwargs = {
        'root_url': out_path,
        'queue'   : urls,
      },
    )
    for p in range(pool)
  ] + [
    multiprocessing.Process(
      target = watchdog,
      kwargs = {
        'total': total,
        'queue': urls,
      }
    )
  ]
  for p in procs:
    p.start()
  for p in procs:
    p.join()
  for p in procs:
    p.terminate()
  return

def check_args(args) -> None:
  """
  Check validity of input arguments
  """
  csv_path = pathlib.Path(args.csv).resolve()
  if not csv_path.exists():
    raise FileNotFoundError(csv_path)
  else:
    args.csv = csv_path

  out_path = pathlib.Path(args.output_dir).resolve()
  if out_path.exists():
    logging.warning("Output path already exists. Is this expected ?")
  else:
    out_path.mkdir(exist_ok = True, parents = True)
  args.output_dir = out_path
  return

def main():
  parser = argparse.ArgumentParser(description = "mine Go programs from GitHub")
  parser.add_argument(
    '-c', '--csv',
    help = "Define path to input csv",
    required = True
  )
  parser.add_argument(
    '-o', '--output_dir',
    help = "Set root output path for cloned repositories",
    required = True
  )
  args = parser.parse_args()
  check_args(args)
  mine(args.csv, args.output_dir)
  return

if __name__ == "__main__":
  main()
  exit()
