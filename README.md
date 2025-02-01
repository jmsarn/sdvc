# Simple Data Version Control

Inspired by [DVC](https://dvc.org/). SDVC is smaller in scope, focusing on managing data
through remote cloud storage (AWS S3) and leaving references to those
files in the Git repository via `*.sdvc` files.

## Important differences from DVC

* SDVC delegates version management to the remote storage
(e.g., bucket versioning for AWS S3). Depending on the implementation
of the remote storage, full copies of each version will be
retained. This is different from DVC which tries to optimize
storage by only storing the diff between versions

## Usage

```bash
# Initialize SDVC in current Git project
# the --remote specifies where SDVC will store files
sdvc init --remote s3://bucket

# Create a reference to a file
sdvc add large-file.parquet

# Upload the file to the remote 
sdvc push large-file.parquet

# Track the reference *.sdvc file with Git
git add large-file.parquet.sdvc
git commit -m "Add large-file.parquet"
git push
```
