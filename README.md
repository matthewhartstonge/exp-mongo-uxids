# exp-mongo-uxids

This experiment attempts to see the difference in storage size (indices+overall) 
in mongo between UUIDs v4 (universally unique identifier, with complete randomness) 
pitted against ULIDs (Universally Unique Lexicographically Sortable Identifier) 
when stored as strings.

First off, clearly UUIDs have a big disadvantage here due to being stored with 
36 characters versus ULID's 26, but critically I wanted to compare how well 
ULIDs scale in indices due to having time-based bits which should lend itself 
to a much higher compression rate.

## Differences
UUIDs can be suboptimal for many use-cases because:

- It isn't the most character efficient way of encoding 128 bits of randomness
- UUID v1/v2 is impractical in many environments, as it requires access to a unique, stable MAC address
- UUID v3/v5 requires a unique seed and produces randomly distributed IDs, which can cause fragmentation in many data structures
- UUID v4 provides no other information than randomness which can cause fragmentation in many data structures

Instead, herein is proposed ULID:

```javascript
ulid() // 01ARZ3NDEKTSV4RRFFQ69G5FAV
```

- 128-bit compatibility with UUID
- 1.21e+24 unique ULIDs per millisecond
- Lexicographically sortable!
- Canonically encoded as a 26 character string, as opposed to the 36 character UUID
- Uses Crockford's base32 for better efficiency and readability (5 bits per character)
- Case insensitive
- No special characters (URL safe)
- Monotonic sort order (correctly detects and handles the same millisecond)

## Storage
So, how do the two compare when it comes to mongo storage size and indices?

Using mongo 3.6 and storing 10,000,000 records with the following structure:

```bson
{
    "_id" : ObjectId("5fae69a6aa5b6100be68876c"),
    "id" : "01EQ0MSCBMD6AAXDFD62JKZ93M"
}
```

## Quick Shootout
Quick overview from 10 million records, where storage size excludes the index 
size, just the raw documents on disk.

| Type   | ID Index Size | Compressed Storage Size | Uncompressed Storage Size |
|--------|---------------|-------------------------|---------------------------|
| UUIDv4 | ~987MBs       | ~508MBs                 | ~638MBs                   |
| ULID   | ~170MBs       | ~258MBs                 | ~543MBs                   |

- Expected = `-27.8%`
- Index Size = `-82.7761%`
- Compressed Storage = `-49.2126%`
- Uncompressed Storage = `-14.8903%`

I suspect that in a production situation, the index size difference wouldn't 
hold up as well as this trivial experiment due to generating 10 million time 
based ids in close succession, but could provide good improvements in high load
systems.

You also gain the resource's created time baked into the ID if you don't want 
to store it in a separate field, as you can calculate it down to the millisecond
by parsing the first 48-bits from the ULID.

UUID index size fluctuated wildly between runs, from 915MBs - 1050MBs, this will
come down to the randomness of UUID generation as the IDs hold no sequential 
nature to them making it harder to bucket. This also makes the id index size
larger than the records stored on disk uncompressed.

ULIDs on the other hand remained constant between runs in index, compressed and
uncompressed storage sizes.

## Full Stats
### UUIDs
`db.getCollection('uuids').stats(1024*1024)`
```json
{
    "ns" : "testIds.uuids",
    "size" : 638,
    "count" : 10000000,
    "avgObjSize" : 67,
    "storageSize" : 508,
    "capped" : false,
    "wiredTiger" : {
        "metadata" : {
            "formatVersion" : 1
        },
        "creationString" : "access_pattern_hint=none,allocation_size=4KB,app_metadata=(formatVersion=1),assert=(commit_timestamp=none,read_timestamp=none),block_allocation=best,block_compressor=snappy,cache_resident=false,checksum=on,colgroups=,collator=,columns=,dictionary=0,encryption=(keyid=,name=),exclusive=false,extractor=,format=btree,huffman_key=,huffman_value=,ignore_in_memory_cache_size=false,immutable=false,internal_item_max=0,internal_key_max=0,internal_key_truncate=true,internal_page_max=4KB,key_format=q,key_gap=10,leaf_item_max=0,leaf_key_max=0,leaf_page_max=32KB,leaf_value_max=64MB,log=(enabled=true),lsm=(auto_throttle=true,bloom=true,bloom_bit_count=16,bloom_config=,bloom_hash_count=8,bloom_oldest=false,chunk_count_limit=0,chunk_max=5GB,chunk_size=10MB,merge_custom=(prefix=,start_generation=0,suffix=),merge_max=15,merge_min=0),memory_page_image_max=0,memory_page_max=10m,os_cache_dirty_max=0,os_cache_max=0,prefix_compression=false,prefix_compression_min=4,source=,split_deepen_min_child=0,split_deepen_per_child=0,split_pct=90,type=file,value_format=u",
        "type" : "file",
        "uri" : "statistics:table:collection-13-2516323338537510223",
        "LSM" : {
            "bloom filter false positives" : 0,
            "bloom filter hits" : 0,
            "bloom filter misses" : 0,
            "bloom filter pages evicted from cache" : 0,
            "bloom filter pages read into cache" : 0,
            "bloom filters in the LSM tree" : 0,
            "chunks in the LSM tree" : 0,
            "highest merge generation in the LSM tree" : 0,
            "queries that could have benefited from a Bloom filter that did not exist" : 0,
            "sleep for LSM checkpoint throttle" : 0,
            "sleep for LSM merge throttle" : 0,
            "total size of bloom filters" : 0
        },
        "block-manager" : {
            "allocations requiring file extension" : 26999,
            "blocks allocated" : 27703,
            "blocks freed" : 863,
            "checkpoint size" : 533299200,
            "file allocation unit size" : 4096,
            "file bytes available for reuse" : 32768,
            "file magic number" : 120897,
            "file major version number" : 1,
            "file size in bytes" : 533348352,
            "minor version number" : 0
        },
        "btree" : {
            "btree checkpoint generation" : 101,
            "column-store fixed-size leaf pages" : 0,
            "column-store internal pages" : 0,
            "column-store variable-size RLE encoded values" : 0,
            "column-store variable-size deleted values" : 0,
            "column-store variable-size leaf pages" : 0,
            "fixed-record size" : 0,
            "maximum internal page key size" : 368,
            "maximum internal page size" : 4096,
            "maximum leaf page key size" : 2867,
            "maximum leaf page size" : 32768,
            "maximum leaf page value size" : 67108864,
            "maximum tree depth" : 4,
            "number of key/value pairs" : 0,
            "overflow pages" : 0,
            "pages rewritten by compaction" : 0,
            "row-store internal pages" : 0,
            "row-store leaf pages" : 0
        },
        "cache" : {
            "bytes currently in the cache" : 279285462,
            "bytes read into cache" : 2981916,
            "bytes written from cache" : 744789001,
            "checkpoint blocked page eviction" : 0,
            "data source pages selected for eviction unable to be evicted" : 808,
            "eviction walk passes of a file" : 8389,
            "eviction walk target pages histogram - 0-9" : 7869,
            "eviction walk target pages histogram - 10-31" : 337,
            "eviction walk target pages histogram - 128 and higher" : 0,
            "eviction walk target pages histogram - 32-63" : 16,
            "eviction walk target pages histogram - 64-128" : 167,
            "eviction walks abandoned" : 18,
            "eviction walks gave up because they restarted their walk twice" : 17,
            "eviction walks gave up because they saw too many pages and found no candidates" : 7606,
            "eviction walks gave up because they saw too many pages and found too few candidates" : 59,
            "eviction walks reached end of tree" : 5108,
            "eviction walks started from root of tree" : 7703,
            "eviction walks started from saved location in tree" : 686,
            "hazard pointer blocked page eviction" : 7,
            "in-memory page passed criteria to be split" : 408,
            "in-memory page splits" : 187,
            "internal pages evicted" : 79,
            "internal pages split during eviction" : 2,
            "leaf pages split during eviction" : 358,
            "modified pages evicted" : 1753,
            "overflow pages read into cache" : 0,
            "page split during eviction deepened the tree" : 1,
            "page written requiring cache overflow records" : 615,
            "pages read into cache" : 781,
            "pages read into cache after truncate" : 1,
            "pages read into cache after truncate in prepare state" : 0,
            "pages read into cache requiring cache overflow entries" : 780,
            "pages requested from the cache" : 16042646,
            "pages seen by eviction walk" : 1854646,
            "pages written from cache" : 27690,
            "pages written requiring in-memory restoration" : 706,
            "tracked dirty bytes in the cache" : 0,
            "unmodified pages evicted" : 6398
        },
        "cache_walk" : {
            "Average difference between current eviction generation when the page was last considered" : 0,
            "Average on-disk page image size seen" : 0,
            "Average time in cache for pages that have been visited by the eviction server" : 0,
            "Average time in cache for pages that have not been visited by the eviction server" : 0,
            "Clean pages currently in cache" : 0,
            "Current eviction generation" : 0,
            "Dirty pages currently in cache" : 0,
            "Entries in the root page" : 0,
            "Internal pages currently in cache" : 0,
            "Leaf pages currently in cache" : 0,
            "Maximum difference between current eviction generation when the page was last considered" : 0,
            "Maximum page size seen" : 0,
            "Minimum on-disk page image size seen" : 0,
            "Number of pages never visited by eviction server" : 0,
            "On-disk page image sizes smaller than a single allocation unit" : 0,
            "Pages created in memory and never written" : 0,
            "Pages currently queued for eviction" : 0,
            "Pages that could not be queued for eviction" : 0,
            "Refs skipped during cache traversal" : 0,
            "Size of the root page" : 0,
            "Total number of pages currently in cache" : 0
        },
        "compression" : {
            "compressed pages read" : 166,
            "compressed pages written" : 26142,
            "page written failed to compress" : 290,
            "page written was too small to compress" : 1254,
            "raw compression call failed, additional data available" : 0,
            "raw compression call failed, no additional data available" : 0,
            "raw compression call succeeded" : 0
        },
        "cursor" : {
            "bulk-loaded cursor-insert calls" : 0,
            "close calls that result in cache" : 0,
            "create calls" : 3,
            "cursor operation restarted" : 0,
            "cursor-insert key and value bytes inserted" : 709917635,
            "cursor-remove key bytes removed" : 0,
            "cursor-update value bytes updated" : 0,
            "cursors reused from cache" : 156299,
            "insert calls" : 10000000,
            "modify calls" : 0,
            "next calls" : 1,
            "open cursor count" : 0,
            "prev calls" : 1,
            "remove calls" : 0,
            "reserve calls" : 0,
            "reset calls" : 312604,
            "search calls" : 0,
            "search near calls" : 0,
            "truncate calls" : 0,
            "update calls" : 0
        },
        "reconciliation" : {
            "dictionary matches" : 0,
            "fast-path pages deleted" : 0,
            "internal page key bytes discarded using suffix compression" : 52453,
            "internal page multi-block writes" : 7,
            "internal-page overflow keys" : 0,
            "leaf page key bytes discarded using prefix compression" : 0,
            "leaf page multi-block writes" : 362,
            "leaf-page overflow keys" : 0,
            "maximum blocks required for a page" : 1,
            "overflow values written" : 0,
            "page checksum matches" : 246,
            "page reconciliation calls" : 4069,
            "page reconciliation calls for eviction" : 3133,
            "pages deleted" : 781
        },
        "session" : {
            "object compaction" : 0
        },
        "transaction" : {
            "update conflicts" : 0
        }
    },
    "nindexes" : 2,
    "totalIndexSize" : 1084,
    "indexSizes" : {
        "_id_" : 97,
        "uuids_id" : 987
    },
    "ok" : 1.0
}
```

### ULIDs
`db.getCollection('ulids').stats(1024*1024)`
```json
{
    "ns" : "testIds.ulids",
    "size" : 543,
    "count" : 10000000,
    "avgObjSize" : 57,
    "storageSize" : 258,
    "capped" : false,
    "wiredTiger" : {
        "metadata" : {
            "formatVersion" : 1
        },
        "creationString" : "access_pattern_hint=none,allocation_size=4KB,app_metadata=(formatVersion=1),assert=(commit_timestamp=none,read_timestamp=none),block_allocation=best,block_compressor=snappy,cache_resident=false,checksum=on,colgroups=,collator=,columns=,dictionary=0,encryption=(keyid=,name=),exclusive=false,extractor=,format=btree,huffman_key=,huffman_value=,ignore_in_memory_cache_size=false,immutable=false,internal_item_max=0,internal_key_max=0,internal_key_truncate=true,internal_page_max=4KB,key_format=q,key_gap=10,leaf_item_max=0,leaf_key_max=0,leaf_page_max=32KB,leaf_value_max=64MB,log=(enabled=true),lsm=(auto_throttle=true,bloom=true,bloom_bit_count=16,bloom_config=,bloom_hash_count=8,bloom_oldest=false,chunk_count_limit=0,chunk_max=5GB,chunk_size=10MB,merge_custom=(prefix=,start_generation=0,suffix=),merge_max=15,merge_min=0),memory_page_image_max=0,memory_page_max=10m,os_cache_dirty_max=0,os_cache_max=0,prefix_compression=false,prefix_compression_min=4,source=,split_deepen_min_child=0,split_deepen_per_child=0,split_pct=90,type=file,value_format=u",
        "type" : "file",
        "uri" : "statistics:table:collection-16-2516323338537510223",
        "LSM" : {
            "bloom filter false positives" : 0,
            "bloom filter hits" : 0,
            "bloom filter misses" : 0,
            "bloom filter pages evicted from cache" : 0,
            "bloom filter pages read into cache" : 0,
            "bloom filters in the LSM tree" : 0,
            "chunks in the LSM tree" : 0,
            "highest merge generation in the LSM tree" : 0,
            "queries that could have benefited from a Bloom filter that did not exist" : 0,
            "sleep for LSM checkpoint throttle" : 0,
            "sleep for LSM merge throttle" : 0,
            "total size of bloom filters" : 0
        },
        "block-manager" : {
            "allocations requiring file extension" : 22286,
            "blocks allocated" : 22298,
            "blocks freed" : 33,
            "checkpoint size" : 271155200,
            "file allocation unit size" : 4096,
            "file bytes available for reuse" : 77824,
            "file magic number" : 120897,
            "file major version number" : 1,
            "file size in bytes" : 271249408,
            "minor version number" : 0
        },
        "btree" : {
            "btree checkpoint generation" : 100,
            "column-store fixed-size leaf pages" : 0,
            "column-store internal pages" : 0,
            "column-store variable-size RLE encoded values" : 0,
            "column-store variable-size deleted values" : 0,
            "column-store variable-size leaf pages" : 0,
            "fixed-record size" : 0,
            "maximum internal page key size" : 368,
            "maximum internal page size" : 4096,
            "maximum leaf page key size" : 2867,
            "maximum leaf page size" : 32768,
            "maximum leaf page value size" : 67108864,
            "maximum tree depth" : 4,
            "number of key/value pairs" : 0,
            "overflow pages" : 0,
            "pages rewritten by compaction" : 0,
            "row-store internal pages" : 0,
            "row-store leaf pages" : 0
        },
        "cache" : {
            "bytes currently in the cache" : 287213683,
            "bytes read into cache" : 28647,
            "bytes written from cache" : 631447384,
            "checkpoint blocked page eviction" : 3,
            "data source pages selected for eviction unable to be evicted" : 50,
            "eviction walk passes of a file" : 502,
            "eviction walk target pages histogram - 0-9" : 155,
            "eviction walk target pages histogram - 10-31" : 63,
            "eviction walk target pages histogram - 128 and higher" : 0,
            "eviction walk target pages histogram - 32-63" : 66,
            "eviction walk target pages histogram - 64-128" : 218,
            "eviction walks abandoned" : 56,
            "eviction walks gave up because they restarted their walk twice" : 10,
            "eviction walks gave up because they saw too many pages and found no candidates" : 81,
            "eviction walks gave up because they saw too many pages and found too few candidates" : 12,
            "eviction walks reached end of tree" : 134,
            "eviction walks started from root of tree" : 159,
            "eviction walks started from saved location in tree" : 343,
            "hazard pointer blocked page eviction" : 0,
            "in-memory page passed criteria to be split" : 384,
            "in-memory page splits" : 148,
            "internal pages evicted" : 3,
            "internal pages split during eviction" : 2,
            "leaf pages split during eviction" : 153,
            "modified pages evicted" : 156,
            "overflow pages read into cache" : 0,
            "page split during eviction deepened the tree" : 1,
            "page written requiring cache overflow records" : 0,
            "pages read into cache" : 1,
            "pages read into cache after truncate" : 1,
            "pages read into cache after truncate in prepare state" : 0,
            "pages read into cache requiring cache overflow entries" : 0,
            "pages requested from the cache" : 15300230,
            "pages seen by eviction walk" : 52861,
            "pages written from cache" : 22293,
            "pages written requiring in-memory restoration" : 1,
            "tracked dirty bytes in the cache" : 0,
            "unmodified pages evicted" : 13260
        },
        "cache_walk" : {
            "Average difference between current eviction generation when the page was last considered" : 0,
            "Average on-disk page image size seen" : 0,
            "Average time in cache for pages that have been visited by the eviction server" : 0,
            "Average time in cache for pages that have not been visited by the eviction server" : 0,
            "Clean pages currently in cache" : 0,
            "Current eviction generation" : 0,
            "Dirty pages currently in cache" : 0,
            "Entries in the root page" : 0,
            "Internal pages currently in cache" : 0,
            "Leaf pages currently in cache" : 0,
            "Maximum difference between current eviction generation when the page was last considered" : 0,
            "Maximum page size seen" : 0,
            "Minimum on-disk page image size seen" : 0,
            "Number of pages never visited by eviction server" : 0,
            "On-disk page image sizes smaller than a single allocation unit" : 0,
            "Pages created in memory and never written" : 0,
            "Pages currently queued for eviction" : 0,
            "Pages that could not be queued for eviction" : 0,
            "Refs skipped during cache traversal" : 0,
            "Size of the root page" : 0,
            "Total number of pages currently in cache" : 0
        },
        "compression" : {
            "compressed pages read" : 1,
            "compressed pages written" : 22060,
            "page written failed to compress" : 0,
            "page written was too small to compress" : 233,
            "raw compression call failed, additional data available" : 0,
            "raw compression call failed, no additional data available" : 0,
            "raw compression call succeeded" : 0
        },
        "cursor" : {
            "bulk-loaded cursor-insert calls" : 0,
            "close calls that result in cache" : 0,
            "create calls" : 3,
            "cursor operation restarted" : 0,
            "cursor-insert key and value bytes inserted" : 609917635,
            "cursor-remove key bytes removed" : 0,
            "cursor-update value bytes updated" : 0,
            "cursors reused from cache" : 156300,
            "insert calls" : 10000000,
            "modify calls" : 0,
            "next calls" : 102,
            "open cursor count" : 0,
            "prev calls" : 1,
            "remove calls" : 0,
            "reserve calls" : 0,
            "reset calls" : 312607,
            "search calls" : 0,
            "search near calls" : 0,
            "truncate calls" : 0,
            "update calls" : 0
        },
        "reconciliation" : {
            "dictionary matches" : 0,
            "fast-path pages deleted" : 0,
            "internal page key bytes discarded using suffix compression" : 44469,
            "internal page multi-block writes" : 4,
            "internal-page overflow keys" : 0,
            "leaf page key bytes discarded using prefix compression" : 0,
            "leaf page multi-block writes" : 159,
            "leaf-page overflow keys" : 0,
            "maximum blocks required for a page" : 1,
            "overflow values written" : 0,
            "page checksum matches" : 175,
            "page reconciliation calls" : 364,
            "page reconciliation calls for eviction" : 149,
            "pages deleted" : 1
        },
        "session" : {
            "object compaction" : 0
        },
        "transaction" : {
            "update conflicts" : 0
        }
    },
    "nindexes" : 2,
    "totalIndexSize" : 266,
    "indexSizes" : {
        "_id_" : 96,
        "ulids_id" : 170
    },
    "ok" : 1.0
}
```
